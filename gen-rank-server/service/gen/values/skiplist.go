package values

import (
	"bytes"
	"fmt"
	"math"
	"math/bits"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	MaxLevel = 25
	//Eps          = 0.00001
	HeadNodeName = "-inf-"
	TailNodeName = "+inf+"
	DebugMode    = false
)

/* 示例数据:
Lv4  -inf
Lv3  -inf     B
Lv2  -inf     B           F
Lv1  -inf  A  B  C  D  E  F  G  H  +inf
指向每一层的后置节点：
B.next[0] = C
B.next[1] = F
B.next[2] = +inf
每一层距离后置节点的跨度：
B.span[0] = 1
B.span[1] = 4
第一层的前置节点：
B.prev = A
A.prev = -inf
+inf.prev = H
*/

type Element struct {
	next [MaxLevel]*Element /*每一层的后置节点*/
	span [MaxLevel]int32    /*每一层距离后置节点的跨度*/
	prev *Element           /*第一层的前置节点*/
	name string             /*唯一名称*/
	val  int64              /*分值，可重复*/
}

func (e *Element) Name() string { return e.name }

func (e *Element) Val() int64 { return e.val }

/*第一层的后置节点*/
func (e *Element) Next() *Element {
	if e.next[0] != nil && e.next[0].name == TailNodeName {
		return nil
	}
	return e.next[0]
}

/*第一层的前置节点*/
func (e *Element) Prev() *Element {
	if e.prev != nil && e.prev.name == HeadNodeName {
		return nil
	}
	return e.prev
}

type SkipList struct {
	headNode *Element         /*头节点，对应 -inf 节点*/
	tailNode *Element         /*尾节点，对应 +inf 节点*/
	maxLevel int              /*当前最大层数*/
	elements map[string]int64 /*元素列表*/
	compare  Compare
	*sync.RWMutex
}

type Compare interface {
	Equal(a, b int64) bool
	Less(a, b int64) bool
}

func NewSkipList(compare Compare) *SkipList {
	return NewSeed(time.Now().UTC().UnixNano(), compare)
}

func NewSeed(seed int64, compare Compare) *SkipList {
	rand.Seed(seed)

	headNode := &Element{
		next: [MaxLevel]*Element{},
		span: [MaxLevel]int32{},
		prev: nil,
		name: HeadNodeName,
		val:  math.MinInt64,
	}

	tailNode := &Element{
		next: [MaxLevel]*Element{},
		span: [MaxLevel]int32{},
		prev: nil,
		name: TailNodeName,
		val:  math.MaxInt64,
	}

	/*头节点的next初始为尾节点，跨度为1 */
	for i := MaxLevel - 1; i >= 0; i-- {
		headNode.next[i] = tailNode
		headNode.span[i] = 1
	}
	tailNode.prev = headNode

	return &SkipList{
		headNode: headNode,
		tailNode: tailNode,
		maxLevel: 0,
		elements: make(map[string]int64),
		compare:  compare,
		RWMutex:  &sync.RWMutex{},
	}
}

/*插入新节点*/
func (t *SkipList) Insert(name string, value int64) {
	t.Lock()
	defer t.Unlock()
	if name == HeadNodeName || name == TailNodeName {
		return
	}
	if t.compare.Equal(t.headNode.val, value) || t.compare.Equal(t.tailNode.val, value) {
		fmt.Println("SkipList Insert error: don't support math.Inf(-1) or math.Inf(1).")
		return
	}

	/*name已存在，若score未变，忽略；若score改变，删除已有节点*/
	if currScore, ok := t.elements[name]; ok {
		if t.compare.Equal(currScore, value) {
			return
		} else {
			t.Delete(name)
		}
	}

	/*生成随机层数； 若超过maxLevel，在maxLevel基础加1，避免跳跃式增长*/
	elemLevel := generateLevel()
	if elemLevel > t.maxLevel {
		elemLevel = t.maxLevel + 1
		t.maxLevel = elemLevel
	}

	elem := &Element{
		next: [MaxLevel]*Element{},
		span: [MaxLevel]int32{},
		prev: nil,
		name: name,
		val:  value,
	}
	t.elements[name] = value

	var (
		index    = t.maxLevel /*从顶层的头节点开始*/
		currNode = t.headNode

		prevs = [MaxLevel]struct { /*记录新元素的每一层的前置节点，以及rank*/
			node *Element
			rank int32
		}{}
	)

	for {
		nextNode := currNode.next[index]

		if !t.compare.Less(nextNode.val, elem.val) { /*找到index层第一个>=新元素的节点*/
			prevs[index].node = currNode

			if index <= elemLevel { /*index层需要插入新节点*/
				elem.next[index] = nextNode
				currNode.next[index] = elem

				if index == 0 { /*第一层更新prev指针*/
					elem.prev = currNode
					nextNode.prev = elem
				}
			}
		}

		if t.compare.Less(nextNode.val, elem.val) { /*尚未找到index层第一个>=新元素的节点*/
			prevs[index].rank += currNode.span[index] /*累加rank*/
			currNode = nextNode                       /*向右遍历*/

		} else {
			if index--; index < 0 { /*转到下一层*/
				break
			} else {
				prevs[index].rank = prevs[index+1].rank /*继承上一层得到的rank*/
			}
		}
	}

	if DebugMode {
		for i, p := range prevs {
			if p.node != nil {
				fmt.Printf("prev, %d, %s, span=%d, rank=%d \n", i, p.node.name, p.node.span, p.rank)
			}
		}
	}

	elemRank := prevs[0].rank + 1
	for i := 0; i <= elemLevel; i++ {
		/* 新元素的span = 前置节点rank + 前置节点的span - 新元素rank + 1 */
		elem.span[i] = prevs[i].rank + prevs[i].node.span[i] - elemRank + 1

		/* 前置节点的span = 新元素rank - 前置节点rank */
		prevs[i].node.span[i] = elemRank - prevs[i].rank
	}

	/* 新元素没有插入的层级，每个前置节点的span加1 */
	for i := elemLevel + 1; i <= t.maxLevel; i++ {
		prevs[i].node.span[i]++
	}

	/* 尚未使用的层级，headNode的span加1 */
	for i := t.maxLevel + 1; i < MaxLevel; i++ {
		t.headNode.span[i]++
	}

	if DebugMode {
		fmt.Printf("elem, %s, span=%d, elemRank=%d, elemLevel=%d\n",
			elem.name, elem.span, elemRank, elemLevel)
		fmt.Println(t.PrintNodes())
	}
}

func (t *SkipList) Find(name string) (foundItem *Element) {
	t.RLock()
	defer t.RUnlock()
	value, ok := t.elements[name]
	if !ok {
		return
	}

	currNode := t.headNode

	for i := t.maxLevel; i >= 0; i-- {

		nextNode := currNode.next[i]
		for t.compare.Less(nextNode.val, value) {
			currNode = nextNode
			nextNode = nextNode.next[i]
		}

		if t.compare.Equal(nextNode.val, value) {
			foundItem = nextNode
			break
		}
	}
	return
}

func (t *SkipList) Delete(name string) {
	t.Lock()
	defer t.Unlock()
	value, ok := t.elements[name]
	if !ok {
		return
	}

	currNode := t.headNode

	for i := t.maxLevel; i >= 0; i-- {

		nextNode := currNode.next[i]
		for t.compare.Less(nextNode.val, value) { /*向右遍历*/
			currNode = nextNode
			nextNode = nextNode.next[i]
		}

		if t.compare.Equal(nextNode.val, value) { /*当前层级，找到待删除节点*/
			delNode := nextNode
			currNode.span[i] += delNode.span[i] - 1 /*前置节点的span增加 */
			currNode.next[i] = delNode.next[i]

			if i == 0 {
				delNode.next[i].prev = currNode
				delete(t.elements, name)
			}

			if t.headNode.next[i] == t.tailNode && i > 0 { /*消除空层*/
				t.maxLevel = i - 1
			}
		} else {
			currNode.span[i]-- /*当前层级没有，前置节点span减1 */
		}
	}

	/*尚未使用的层级，headNode的span减1 */
	for i := t.maxLevel + 1; i < MaxLevel; i++ {
		t.headNode.span[i]--
	}
}

func (t *SkipList) GetRank(name string) (rank int, exist bool) {
	t.RLock()
	defer t.RUnlock()
	value, ok := t.elements[name]
	if !ok {
		return
	}

	currNode := t.headNode
	var elemRank int32

	for i := t.maxLevel; i >= 0; i-- {

		nextNode := currNode.next[i]
		for t.compare.Less(nextNode.val, value) {
			elemRank += currNode.span[i]

			currNode = nextNode
			nextNode = nextNode.next[i]
		}

		if t.compare.Equal(nextNode.val, value) {
			rank = int(elemRank + currNode.span[i])
			exist = true
			break
		}
	}
	return
}

func (t *SkipList) FindByRank(rank int) (foundItem *Element) {
	t.RLock()
	defer t.RUnlock()
	if rank < 1 || rank > len(t.elements) {
		return nil
	}

	currNode := t.headNode
	var elemRank int32

	for i := t.maxLevel; i >= 0; i-- {

		for currNode.next[i] != t.tailNode {

			if nextRank := elemRank + currNode.span[i]; int(nextRank) <= rank {
				elemRank = nextRank
				currNode = currNode.next[i]
			} else {
				break
			}
		}

		if int(elemRank) == rank {
			foundItem = currNode
			break
		}
	}
	return
}

func (t *SkipList) IsEmpty() bool {
	t.RLock()
	defer t.RUnlock()
	return t.headNode.next[0] == t.tailNode
}

func (t *SkipList) GetNodeCount() int {
	t.RLock()
	defer t.RUnlock()
	return len(t.elements)
}

func (t *SkipList) Get(name string) (int64, bool) {
	t.RLock()
	defer t.RUnlock()
	value, ok := t.elements[name]
	return value, ok
}

func (t *SkipList) PrintNodes() string {
	t.RLock()
	defer t.RUnlock()
	levels := make([]string, 0, MaxLevel)
	var buff bytes.Buffer

	for i := t.maxLevel; i >= 0; i-- {
		buff.Reset()
		buff.WriteString("[" + strconv.Itoa(i) + "] ")

		for node := t.headNode; node != nil; node = node.next[i] {
			buff.WriteString(node.name)
			buff.WriteString(fmt.Sprintf(" (%d) ", node.span[i]))
		}

		levels = append(levels, buff.String())
	}

	return strings.Join(levels, "\n")
}

func (t *SkipList) PrintLevels() string {
	t.RLock()
	defer t.RUnlock()
	levels := make([]string, 0, MaxLevel)
	wholeCount := 0

	for i := t.maxLevel; i >= 0; i-- {
		count := 0
		for node := t.headNode.next[i]; node != t.tailNode; node = node.next[i] {
			count++
		}

		levels = append(levels, fmt.Sprintf("[%02d] %d", i, count))
		wholeCount += count
	}

	levels = append(levels, "whole count="+strconv.Itoa(wholeCount))
	return strings.Join(levels, "\n")
}

/*Return random layers*/
func generateLevel() int {
	var x = rand.Uint64() & ((1 << uint(MaxLevel-1)) - 1) /*Random value x, bit number < MAX_LEVEL*/
	zeroes := bits.TrailingZeros64(x)                     /*Starting from the tail, the number of bits 0*/

	level := MaxLevel - 1
	if zeroes < MaxLevel {
		level = zeroes
	}
	return level
}
