#bytespool

全局化bytespool

目前支持这些大小
```go
// 小数间隔的目的是为了更高概率复用。但是可能会造成内存使用偏高一点。对gc友好
var sizes = []int{
	1 << 6,
	1 << 8,
	1 << 10,
	1 << 12,
	1 << 14,
	1 << 15,
	1 << 16,
	1 << 17,
	1 << 18,
	1 << 19,
	1 << 20,
	1 << 21,
	1 << 22,
	1 << 23,
	1 << 24,
}
```

使用:
```go
data:=bytespool.Get(1) // len(data) 不确定. cap(data)==64
// 如果append, 需要手动 data=data[:0]
// 如果直接赋值或者copy 需要手动 data=data[:n]
data:=bytespool.Get(64) // len(data) 不确定. cap(data)==64

data:==bytespool.Get(1<<7)// cap(data)=1<<8
// 使用完之后
bytespool.Put(data)
```
