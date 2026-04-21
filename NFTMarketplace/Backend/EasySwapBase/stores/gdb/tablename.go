package gdb

import (
	"github.com/ProjectsTask/EasySwapBase/stores/gdb/orderbookmodel/multi"
)

func GetMultiProjectOrderTableName(project string, chain string) string {
	if project == OrderBookDexProject {
		return multi.OrderTableName(chain)
	} else {
		return ""
	}

}

// GetMultiProjectItemTableName 根据传入的项目名称和区块链名称，获取对应的物品表名。
// 参数:
//   project - 项目名称，用于判断是否为目标项目。
//   chain - 区块链名称，用于生成特定区块链下的表名。
// 返回值:
//   如果项目名称为 OrderBookDexProject，则返回对应区块链的物品表名；
//   否则，返回空字符串，表示不支持该项目或使用默认表名。
func GetMultiProjectItemTableName(project string, chain string) string {
    // 检查传入的项目名称是否为 OrderBookDexProject
    if project == OrderBookDexProject {
        // 若项目名称匹配，调用 multi 包的 ItemTableName 函数，传入链名，获取对应表名
        return multi.ItemTableName(chain)
    } else {
        // 若项目名称不匹配，返回空字符串
        return ""
    }
}

func GetMultiProjectCollectionTableName(project string, chain string) string {
	if project == OrderBookDexProject {
		return multi.CollectionTableName(chain)
	} else {
		return ""
	}
}

func GetMultiProjectActivityTableName(project string, chain string) string {
	if project == OrderBookDexProject {
		return multi.ActivityTableName(chain)
	} else {
		return ""
	}
}

func GetMultiProjectCollectionFloorPriceTableName(project string, chain string) string {
	if project == OrderBookDexProject {
		return multi.CollectionFloorPriceTableName(chain)
	} else {
		return ""
	}
}

func GetMultiProjectItemExternalTableName(project string, chain string) string {
	if project == OrderBookDexProject {
		return multi.ItemExternalTableName(chain)
	} else {
		return ""
	}
}

func GetMultiProjectItemTraitTableName(project string, chain string) string {
	if project == OrderBookDexProject {
		return multi.ItemTraitTableName(chain)
	} else {
		return ""
	}
}

func GetMultiProjectCollectionTradeTableName(project string, chain string) string {
	if project == OrderBookDexProject {
		return multi.CollectionTradeTableName(chain)
	} else {
		return ""
	}
}
