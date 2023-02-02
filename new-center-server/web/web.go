package web

import (
	"coin-server/common/errmsg"
	"coin-server/common/logger"
	_ "coin-server/new-center-server/env"
	"fmt"

	"net/http"
	"strconv"

	modelspb "coin-server/common/proto/models"
	centerpb "coin-server/common/proto/newcenter"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"go.uber.org/zap"
)

func StartWeb(log *logger.Logger) {
	http.HandleFunc("/tree", httpserver)
	http.ListenAndServe(":9091", nil)
}

func httpserver(w http.ResponseWriter, _ *http.Request) {
	width, height, datas, err := GetData()
	if err != nil {
		fmt.Fprintln(w, zap.Any("err message", err))
		return
	}
	tree := charts.NewTree()

	widthPx := "2100px"
	if width*600 < 2100 {
		widthPx = strconv.FormatInt(width*600, 10) + "px"
	}

	tree.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Width: widthPx, Height: strconv.FormatInt(height*60, 10) + "px"}),
		charts.WithTitleOpts(opts.Title{
			Title: "CenterServer View",
			Left:  "center",
		}),
		charts.WithTooltipOpts(opts.Tooltip{Show: false}),
	)
	tree.AddSeries("tree", *datas).
		SetSeriesOptions(
			charts.WithTreeOpts(
				opts.TreeChart{
					Layout:           "orthogonal",
					Orient:           "LR",
					Roam:             true,
					InitialTreeDepth: 2,
					Leaves: &opts.TreeLeaves{
						Label: &opts.Label{
							Show:     true,
							Position: "right",
							Color:    "Black",
						},
					},
					Top:    "1%",
					Left:   "10%",
					Bottom: "1%",
					Right:  "20%",
				},
			),
			charts.WithLabelOpts(opts.Label{Show: true, Position: "top", Color: "Black"}),
		)
	tree.Render(w)
}

func GetData() (int64, int64, *[]opts.TreeData, *errmsg.ErrMsg) {
	datas, err := GetTreeData()
	if err != nil {
		return 0, 0, nil, err
	}

	maxWidth, maxHight := int64(0), int64(0)
	for _, data := range *datas {
		width, hight := Calculation(1, &data)
		if width > maxWidth {
			maxWidth = width
		}
		if hight > maxHight {
			maxHight = hight
		}
	}
	return maxWidth, maxHight, datas, nil
}

func Calculation(layer int64, data *opts.TreeData) (int64, int64) {
	maxWidth, maxHight := layer, int64(0)
	treeLen := len(data.Children)
	if treeLen == 0 {
		return 0, 0
	}
	maxWidth = layer + 1
	maxHight = int64(treeLen)

	for _, treeData := range data.Children {
		nextWidth, nextHight := Calculation(layer+1, treeData)
		if nextWidth > maxWidth {
			maxWidth = nextWidth
		}
		maxHight += nextHight
	}
	return maxWidth, maxHight
}

func GetTreeData() (*[]opts.TreeData, *errmsg.ErrMsg) {
	veiwData, err := GetEdgeViewData()
	if err != nil {
		return nil, err
	}

	treeData := &opts.TreeData{
		Name: "CenterServer",
	}

	treeData.Children = ProcEdegViewPb(veiwData.Edges)

	return &[]opts.TreeData{
		*treeData,
	}, nil
}

func ProcEdegViewPb(edges []*centerpb.NewCenter_EdgeNode) []*opts.TreeData {
	var treeData []*opts.TreeData
	for _, edgeNode := range edges {
		treeData = append(treeData, &opts.TreeData{
			Name: "edgeNode:" + edgeNode.Addr,
			Children: func(edgeNode *centerpb.NewCenter_EdgeNode) []*opts.TreeData {
				var treeNode []*opts.TreeData
				treeNode = append(treeNode, &opts.TreeData{
					Name: "Attr : " + edgeNode.Addr +
						"\nWeight : " + strconv.FormatInt(edgeNode.Weight, 10) +
						"\nEdgeType : " + strconv.FormatInt(int64(edgeNode.Typ), 10),
				})
				treeNode = append(treeNode, &opts.TreeData{
					Name: "Battles:" + strconv.FormatInt(int64(len(edgeNode.Battles)), 10),
					Children: func(edgeBattlesPb []*centerpb.NewCenter_EdgeBattle) []*opts.TreeData {
						var edgeBattle []*opts.TreeData
						for _, eBattle := range edgeBattlesPb {
							edgeBattle = append(edgeBattle, &opts.TreeData{
								Name: "BattleId  : " + strconv.FormatInt(int64(eBattle.BattleId), 10) +
									"\nMapID  : " + strconv.FormatInt(int64(eBattle.MapId), 10) +
									"\nMapName : " + eBattle.MapName +
									"\nCurNum  : " + strconv.FormatInt(int64(eBattle.CurNum), 10),
								Collapsed: true,
							})
						}
						return edgeBattle
					}(edgeNode.Battles),
				})
				return treeNode
			}(edgeNode),
		})
	}
	return treeData
}

func ProcEdegViewPb2(edges []*centerpb.NewCenter_EdgeNode) []*opts.TreeData {
	var treeData []*opts.TreeData
	for _, edgeNode := range edges {
		treeData = append(treeData, &opts.TreeData{
			Name: "edgeNode:" + edgeNode.Addr,
			Children: func(edgeNode *centerpb.NewCenter_EdgeNode) []*opts.TreeData {
				var treeNode []*opts.TreeData
				treeNode = append(treeNode, &opts.TreeData{
					Name: "Attr:" + edgeNode.Addr,
				})
				treeNode = append(treeNode, &opts.TreeData{
					Name: "Weight:" + strconv.FormatInt(edgeNode.Weight, 10),
				})
				treeNode = append(treeNode, &opts.TreeData{
					Name: "EdgeType:" + strconv.FormatInt(int64(edgeNode.Typ), 10),
				})
				treeNode = append(treeNode, &opts.TreeData{
					Name: "Battles:" + strconv.FormatInt(int64(len(edgeNode.Battles)), 10),
					Children: func(edgeBattlesPb []*centerpb.NewCenter_EdgeBattle) []*opts.TreeData {
						var edgeBattle []*opts.TreeData
						for _, eBattle := range edgeBattlesPb {
							edgeBattle = append(edgeBattle, &opts.TreeData{
								Name: "BattleId  : " + strconv.FormatInt(int64(eBattle.BattleId), 10),
								Children: func(eBattlesPb *centerpb.NewCenter_EdgeBattle) []*opts.TreeData {
									var eBattle []*opts.TreeData
									eBattle = append(eBattle, &opts.TreeData{
										Name: "BattleId  : " + strconv.FormatInt(int64(eBattlesPb.BattleId), 10) +
											"\nMapID  : " + strconv.FormatInt(int64(eBattlesPb.MapId), 10) +
											"\nMapName : " + eBattlesPb.MapName +
											"\nCurNum  : " + strconv.FormatInt(int64(eBattlesPb.CurNum), 10),
										Collapsed: true,
									})
									// eBattle = append(eBattle, &opts.TreeData{
									// 	Name: "BattleId:" + strconv.FormatInt(int64(eBattlesPb.BattleId), 10),
									// })
									// eBattle = append(eBattle, &opts.TreeData{
									// 	Name: "MapID:" + strconv.FormatInt(int64(eBattlesPb.MapId), 10),
									// })
									// eBattle = append(eBattle, &opts.TreeData{
									// 	Name: "MapName:" + eBattlesPb.MapName,
									// })
									// eBattle = append(eBattle, &opts.TreeData{
									// 	Name: "CurNum:" + strconv.FormatInt(int64(eBattlesPb.CurNum), 10),
									// })
									return eBattle
								}(eBattle),
							})
						}
						return edgeBattle
					}(edgeNode.Battles),
				})
				return treeNode
			}(edgeNode),
		})
	}
	return treeData
}

func GetEdgeViewData() (*centerpb.NewCenter_EdgeViewInfo, *errmsg.ErrMsg) {
	return GetTestData()
}

func GetTestData() (*centerpb.NewCenter_EdgeViewInfo, *errmsg.ErrMsg) {
	return &centerpb.NewCenter_EdgeViewInfo{
		Edges: []*centerpb.NewCenter_EdgeNode{
			{
				Addr:   "10.10.10.1",
				Weight: 10,
				Typ:    modelspb.EdgeType_StatelessServer,
				Battles: []*centerpb.NewCenter_EdgeBattle{
					{
						BattleId: 10000,
						MapId:    120001,
						MapName:  "挂机1",
						CurNum:   100,
					},
					{
						BattleId: 10001,
						MapId:    120002,
						MapName:  "挂机2",
						CurNum:   200,
					},
					{
						BattleId: 10002,
						MapId:    120002,
						MapName:  "挂机3",
						CurNum:   300,
					},
				},
			},
			{
				Addr:   "10.10.10.2",
				Weight: 20,
				Typ:    modelspb.EdgeType_StatelessServer,
				Battles: []*centerpb.NewCenter_EdgeBattle{
					{
						BattleId: 20000,
						MapId:    220001,
						MapName:  "2挂机1",
						CurNum:   200,
					},
					{
						BattleId: 20001,
						MapId:    220002,
						MapName:  "2挂机2",
						CurNum:   300,
					},
					{
						BattleId: 20002,
						MapId:    220002,
						MapName:  "2挂机3",
						CurNum:   400,
					},
				},
			},
			{
				Addr:   "10.10.10.2",
				Weight: 20,
				Typ:    modelspb.EdgeType_StatelessServer,
				Battles: []*centerpb.NewCenter_EdgeBattle{
					{
						BattleId: 20000,
						MapId:    220001,
						MapName:  "2挂机1",
						CurNum:   200,
					},
					{
						BattleId: 20001,
						MapId:    220002,
						MapName:  "2挂机2",
						CurNum:   300,
					},
					{
						BattleId: 20002,
						MapId:    220002,
						MapName:  "2挂机3",
						CurNum:   400,
					},
				},
			},
			{
				Addr:   "10.10.10.2",
				Weight: 20,
				Typ:    modelspb.EdgeType_StatelessServer,
				Battles: []*centerpb.NewCenter_EdgeBattle{
					{
						BattleId: 20000,
						MapId:    220001,
						MapName:  "2挂机1",
						CurNum:   200,
					},
					{
						BattleId: 20001,
						MapId:    220002,
						MapName:  "2挂机2",
						CurNum:   300,
					},
					{
						BattleId: 20002,
						MapId:    220002,
						MapName:  "2挂机3",
						CurNum:   400,
					},
				},
			},
			{
				Addr:   "10.10.10.2",
				Weight: 20,
				Typ:    modelspb.EdgeType_StatelessServer,
				Battles: []*centerpb.NewCenter_EdgeBattle{
					{
						BattleId: 20000,
						MapId:    220001,
						MapName:  "2挂机1",
						CurNum:   200,
					},
					{
						BattleId: 20001,
						MapId:    220002,
						MapName:  "2挂机2",
						CurNum:   300,
					},
					{
						BattleId: 20002,
						MapId:    220002,
						MapName:  "2挂机3",
						CurNum:   400,
					},
				},
			},
		},
	}, nil
}
