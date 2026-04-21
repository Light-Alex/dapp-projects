// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import {Initializable} from "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";

// import {ReentrancyGuardUpgradeable} from "@openzeppelin/contracts-upgradeable/utils/ReentrancyGuardUpgradeable.sol";
// 实现红黑树数据结构，用于高效管理价格排序
import {RedBlackTreeLibrary, Price} from "./libraries/RedBlackTreeLibrary.sol";
//LibOrder：定义订单相关的结构和工具函数
// OrderKey：订单的唯一标识符
import {LibOrder, OrderKey} from "./libraries/LibOrder.sol";

error CannotInsertDuplicateOrder(OrderKey orderKey);

// OrderStorage 是一个用于管理订单存储和排序的合约，支持订单的添加、移除、查询和排序功能。
// 它是 EasySwapOrderBook 的核心模块之一，负责高效地管理订单数据
contract OrderStorage is Initializable {
    // Initializable：支持合约的初始化功能，通常用于可升级合约
    // 红黑树数据结构，用于高效管理价格排序
    using RedBlackTreeLibrary for RedBlackTreeLibrary.Tree;

    /// @dev all order keys are wrapped in a sentinel value to avoid collisions
    //存储订单的详细信息，键为订单的唯一标识符
    mapping(OrderKey => LibOrder.DBOrder) public orders;

    /// @dev price tree for each collection and side, sorted by price
    // 每个 NFT 集合和订单类型（买单或卖单）对应的价格树，按价格排序
    mapping(address => mapping(LibOrder.Side => RedBlackTreeLibrary.Tree))
        public priceTrees;

    /// @dev order queue for each collection, side and expecially price, sorted by orderKey
    // 每个 NFT 集合、订单类型和价格对应的订单队列，按订单的插入顺序排序
    // LibOrder.OrderQueue 链表结构，用于管理同一价格下的订单队列
    mapping(address => mapping(LibOrder.Side => mapping(Price => LibOrder.OrderQueue)))
        public orderQueues;

    //初始化合约，通常在可升级合约的构造函数中调用
    function __OrderStorage_init() internal onlyInitializing {}

    function __OrderStorage_init_unchained() internal onlyInitializing {}

    function onePlus(uint256 x) internal pure returns (uint256) {
        unchecked {
            return 1 + x;
        }
    }

    // 获取指定 NFT 集合和订单类型（买单或卖单）的最佳价格的方法
    function getBestPrice(
        address collection, //NFT 集合的合约地址
        LibOrder.Side side //订单类型（买单或卖单）
    ) public view returns (Price price) {
        //LibOrder.Side.Bid：买单
        // LibOrder.Side.List：卖单
        //价格树查询：通过 priceTrees[collection][side] 获取指定集合和订单类型的价格树
        price = (side == LibOrder.Side.Bid)
            ? priceTrees[collection][side].last() // 如果是买单，返回最高价
            : priceTrees[collection][side].first(); // 如果是卖单，返回最低价
    }

    // 获取比当前价格更好的下一个价格。对于买单（Bid），返回比当前价格低的下一个价格；
    // 对于卖单（List），返回比当前价格高的下一个价格
    function getNextBestPrice(
        address collection, //NFT 集合的合约地址
        LibOrder.Side side, //订单类型（买单或卖单）
        Price price //当前价格
    ) public view returns (Price nextBestPrice) {
        if (RedBlackTreeLibrary.isEmpty(price)) {
            //如果输入价格为空
            nextBestPrice = (side == LibOrder.Side.Bid)
                ? priceTrees[collection][side].last() //买单：返回价格树中的最高价
                : priceTrees[collection][side].first(); //卖单：返回价格树中的最低价
        } else {
            // 输入价格非空
            nextBestPrice = (side == LibOrder.Side.Bid)
                ? priceTrees[collection][side].prev(price) //买单：返回价格树中比当前价格低的下一个价格
                : priceTrees[collection][side].next(price); //卖单：返回价格树中比当前价格高的下一个价格
        }
    }

    // 添加一个新订单到存储中
    function _addOrder(
        LibOrder.Order memory order
    ) internal returns (OrderKey orderKey) {
        // 获取订单的hash值唯一标识符
        orderKey = LibOrder.hash(order);
        //  判断订单是否已经存在
        if (orders[orderKey].order.maker != address(0)) {
            revert CannotInsertDuplicateOrder(orderKey);
        }

        // insert price to price tree if not exists
        RedBlackTreeLibrary.Tree storage priceTree = priceTrees[
            order.nft.collection
        ][order.side];
        if (!priceTree.exists(order.price)) {
            // 插入价格到价格树
            priceTree.insert(order.price);
        }

        // insert order to order queue
        LibOrder.OrderQueue storage orderQueue = orderQueues[
            order.nft.collection
        ][order.side][order.price];

        if (LibOrder.isSentinel(orderQueue.head)) {
            // 队列是否初始化
            orderQueues[order.nft.collection][order.side][ // 创建新的队列
                order.price
            ] = LibOrder.OrderQueue(
                LibOrder.ORDERKEY_SENTINEL, // 头位置
                LibOrder.ORDERKEY_SENTINEL // 结尾
            );
            orderQueue = orderQueues[order.nft.collection][order.side][
                order.price
            ];
        }
        if (LibOrder.isSentinel(orderQueue.tail)) {
            // 队列是否为空
            orderQueue.head = orderKey;
            orderQueue.tail = orderKey;
            orders[orderKey] = LibOrder.DBOrder( // 创建新的订单，插入队列， 下一个订单为sentinel
                order,
                LibOrder.ORDERKEY_SENTINEL
            );
        } else {
            // 队列不为空
            orders[orderQueue.tail].next = orderKey; // 将新订单插入队列尾部
            orders[orderKey] = LibOrder.DBOrder(
                order,
                LibOrder.ORDERKEY_SENTINEL
            );
            orderQueue.tail = orderKey;
        }
    }

    //从存储中移除一个订单，遍历订单队列，找到匹配的订单并更新队列的指针关系，同时更新价格树。
    function _removeOrder(
        LibOrder.Order memory order //要移除的订单信息
    ) internal returns (OrderKey orderKey) {
        //orderKey被移除订单的唯一标识符
        //根据订单的 NFT 集合、类型和价格获取对应的订单队列
        LibOrder.OrderQueue storage orderQueue = orderQueues[
            order.nft.collection
        ][order.side][order.price];
        //从队列头部开始遍历
        orderKey = orderQueue.head;
        //记录前一个订单的标识符
        OrderKey prevOrderKey;
        //标记是否找到目标订单
        bool found;
        //遍历订单队列查找目标订单
        while (LibOrder.isNotSentinel(orderKey) && !found) {
            LibOrder.DBOrder memory dbOrder = orders[orderKey];
            if (
                (dbOrder.order.maker == order.maker) &&
                (dbOrder.order.saleKind == order.saleKind) &&
                (dbOrder.order.expiry == order.expiry) &&
                (dbOrder.order.salt == order.salt) &&
                (dbOrder.order.nft.tokenId == order.nft.tokenId) &&
                (dbOrder.order.nft.amount == order.nft.amount)
            ) {
                // 找到匹配的订单
                OrderKey temp = orderKey;
                // emit OrderRemoved(order.nft.collection, orderKey, order.maker, order.side, order.price, order.nft, block.timestamp);
                if (
                    OrderKey.unwrap(orderQueue.head) ==
                    OrderKey.unwrap(orderKey)
                ) {
                    // 如果是队列头部
                    orderQueue.head = dbOrder.next;
                } else {
                    // 如果是队列中间
                    orders[prevOrderKey].next = dbOrder.next;
                }
                if (
                    OrderKey.unwrap(orderQueue.tail) ==
                    OrderKey.unwrap(orderKey)
                ) {
                    // 如果是队列尾部
                    orderQueue.tail = prevOrderKey;
                }
                prevOrderKey = orderKey;
                orderKey = dbOrder.next;
                delete orders[temp];
                found = true;
            } else {
                prevOrderKey = orderKey;
                orderKey = dbOrder.next;
            }
        }
        if (found) {
            // 删除订单并清理价格树
            if (LibOrder.isSentinel(orderQueue.head)) {
                // 如果队列为空，删除队列
                delete orderQueues[order.nft.collection][order.side][
                    order.price
                ];
                // 从价格树中移除价格
                RedBlackTreeLibrary.Tree storage priceTree = priceTrees[
                    order.nft.collection
                ][order.side];
                if (priceTree.exists(order.price)) {
                    priceTree.remove(order.price);
                }
            }
        } else {
            revert("Cannot remove missing order");
        }
    }

    /**
     * @dev Retrieves a list of orders that match the specified criteria.
     * @param collection The address of the NFT collection.
     * @param tokenId The ID of the NFT.
     * @param side The side of the orders to retrieve (buy or sell).
     * @param saleKind The type of sale (fixed price or auction).
     * @param count The maximum number of orders to retrieve.
     * @param price The maximum price of the orders to retrieve.
     * @param firstOrderKey The key of the first order to retrieve.
     * @return resultOrders An array of orders that match the specified criteria.
     * @return nextOrderKey The key of the next order to retrieve.
     */
    // 根据指定条件查询订单列表
    function getOrders(
        address collection, // NFT 集合地址
        uint256 tokenId, // NFT 的 ID
        LibOrder.Side side, // 订单类型（买单或卖单）
        LibOrder.SaleKind saleKind, // 销售类型
        uint256 count, // 最大返回订单数量
        Price price, // 价格过滤条件
        OrderKey firstOrderKey // 起始订单的 Key
    )
        external
        view
        returns (
            // resultOrders 符合条件的订单数组，nextOrderKey 下一个订单的 Key
            LibOrder.Order[] memory resultOrders,
            OrderKey nextOrderKey
        )
    {
        //初始化返回值数组
        resultOrders = new LibOrder.Order[](count);

        if (RedBlackTreeLibrary.isEmpty(price)) {
            // 如果没有指定价格, 获取最佳价格
            price = getBestPrice(collection, side);
        } else {
            // 如果指定了价格且是新的查询（无 firstOrderKey），获取下一个最佳价格
            if (LibOrder.isSentinel(firstOrderKey)) {
                price = getNextBestPrice(collection, side, price);
            }
        }
        // 遍历订单队列
        uint256 i;
        while (RedBlackTreeLibrary.isNotEmpty(price) && i < count) {
            LibOrder.OrderQueue memory orderQueue = orderQueues[collection][
                side
            ][price];
            OrderKey orderKey = orderQueue.head;
            if (LibOrder.isNotSentinel(firstOrderKey)) {
                //如果提供了 firstOrderKey，从该订单开始查询
                while (
                    LibOrder.isNotSentinel(orderKey) &&
                    OrderKey.unwrap(orderKey) != OrderKey.unwrap(firstOrderKey)
                ) {
                    LibOrder.DBOrder memory order = orders[orderKey];
                    orderKey = order.next;
                }
                // 查找到起始订单后，重置 firstOrderKey
                firstOrderKey = LibOrder.ORDERKEY_SENTINEL;
            }

            while (LibOrder.isNotSentinel(orderKey) && i < count) {
                LibOrder.DBOrder memory dbOrder = orders[orderKey];
                orderKey = dbOrder.next;
                // 过期订单过滤
                if (
                    (dbOrder.order.expiry != 0 &&
                        dbOrder.order.expiry < block.timestamp)
                ) {
                    continue;
                }
                // 订单类型匹配
                if (
                    (side == LibOrder.Side.Bid) &&
                    (saleKind == LibOrder.SaleKind.FixedPriceForCollection)
                ) {
                    if (
                        (dbOrder.order.side == LibOrder.Side.Bid) &&
                        (dbOrder.order.saleKind ==
                            LibOrder.SaleKind.FixedPriceForItem)
                    ) {
                        continue;
                    }
                }
                // tokenId 匹配
                if (
                    (side == LibOrder.Side.Bid) &&
                    (saleKind == LibOrder.SaleKind.FixedPriceForItem)
                ) {
                    if (
                        (dbOrder.order.side == LibOrder.Side.Bid) &&
                        (dbOrder.order.saleKind ==
                            LibOrder.SaleKind.FixedPriceForItem) &&
                        (tokenId != dbOrder.order.nft.tokenId)
                    ) {
                        continue;
                    }
                }
                // 添加符合条件的订单
                resultOrders[i] = dbOrder.order;
                nextOrderKey = dbOrder.next;
                i = onePlus(i);
            }
            price = getNextBestPrice(collection, side, price);
        }
    }

    // 查询符合条件的最佳订单
    function getBestOrder(
        address collection, // NFT 集合地址
        uint256 tokenId, // NFT 的 ID
        LibOrder.Side side, // 订单类型（买单或卖单）
        LibOrder.SaleKind saleKind // 销售类型
    ) external view returns (LibOrder.Order memory orderResult) {
        // 获取最佳价格，买单返回最高价，卖单返回最低价
        Price price = getBestPrice(collection, side);
        // 遍历价格树中的每个价格
        while (RedBlackTreeLibrary.isNotEmpty(price)) {
            // 获取每个价格对应的订单队列
            LibOrder.OrderQueue memory orderQueue = orderQueues[collection][
                side
            ][price];
            OrderKey orderKey = orderQueue.head;
            // 遍遍历当前价格的订单队列
            while (LibOrder.isNotSentinel(orderKey)) {
                //使用 orderKey 获取订单的详细信息
                LibOrder.DBOrder memory dbOrder = orders[orderKey];
                //订单过滤
                if (
                    (side == LibOrder.Side.Bid) &&
                    (saleKind == LibOrder.SaleKind.FixedPriceForItem)
                ) {
                    //如果是买单且销售类型为单个 NFT 的固定价格
                    //检查 tokenId 是否匹配。
                    //如果不匹配，跳过当前订单
                    if (
                        (dbOrder.order.side == LibOrder.Side.Bid) &&
                        (dbOrder.order.saleKind ==
                            LibOrder.SaleKind.FixedPriceForItem) &&
                        (tokenId != dbOrder.order.nft.tokenId)
                    ) {
                        orderKey = dbOrder.next;
                        continue;
                    }
                }

                if (
                    (side == LibOrder.Side.Bid) &&
                    (saleKind == LibOrder.SaleKind.FixedPriceForCollection)
                ) {
                    //如果是买单且销售类型为集合的固定价格
                    //跳过销售类型为单个 NFT 的订单
                    if (
                        (dbOrder.order.side == LibOrder.Side.Bid) &&
                        (dbOrder.order.saleKind ==
                            LibOrder.SaleKind.FixedPriceForItem)
                    ) {
                        orderKey = dbOrder.next;
                        continue;
                    }
                }
                //检查订单是否已过期
                if (
                    (dbOrder.order.expiry == 0 ||
                        dbOrder.order.expiry > block.timestamp)
                ) {
                    // 如果 expiry == 0，表示订单永不过期
                    // 如果 expiry > block.timestamp，表示订单未过期
                    // 如果订单未过期，将其设置为结果订单并退出循环
                    orderResult = dbOrder.order;
                    break;
                }
                orderKey = dbOrder.next;
            }
            //如果找到符合条件的订单，退出价格树的遍历
            if (Price.unwrap(orderResult.price) > 0) {
                break;
            }
            // 获取指定价格的下一个最佳价格
            price = getNextBestPrice(collection, side, price);
        }
    }

    uint256[50] private __gap;
}
