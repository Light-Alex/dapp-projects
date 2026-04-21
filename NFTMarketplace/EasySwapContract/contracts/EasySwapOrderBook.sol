// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import {Initializable} from "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import {ContextUpgradeable} from "@openzeppelin/contracts-upgradeable/utils/ContextUpgradeable.sol";
import {OwnableUpgradeable} from "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import {ReentrancyGuardUpgradeable} from "@openzeppelin/contracts-upgradeable/utils/ReentrancyGuardUpgradeable.sol";
import {PausableUpgradeable} from "@openzeppelin/contracts-upgradeable/utils/PausableUpgradeable.sol";

import {LibTransferSafeUpgradeable, IERC721} from "./libraries/LibTransferSafeUpgradeable.sol";
import {Price} from "./libraries/RedBlackTreeLibrary.sol";
import {LibOrder, OrderKey} from "./libraries/LibOrder.sol";
import {LibPayInfo} from "./libraries/LibPayInfo.sol";

import {IEasySwapOrderBook} from "./interface/IEasySwapOrderBook.sol";
import {IEasySwapVault} from "./interface/IEasySwapVault.sol";

import {OrderStorage} from "./OrderStorage.sol";
import {OrderValidator} from "./OrderValidator.sol";
import {ProtocolManager} from "./ProtocolManager.sol";

// NFT 交易市场的订单簿
contract EasySwapOrderBook is
    IEasySwapOrderBook, // 接口
    Initializable, // 可升级合约初始化
    ContextUpgradeable, // 上下文管理
    OwnableUpgradeable, // 所有权管理
    ReentrancyGuardUpgradeable, // 重入保护
    PausableUpgradeable, // 可暂停功能
    OrderStorage, // 订单存储
    ProtocolManager, // 协议管理
    OrderValidator // 订单验证
{
    using LibTransferSafeUpgradeable for address;
    using LibTransferSafeUpgradeable for IERC721;
    // 创建订单事件
    event LogMake(
        OrderKey orderKey,
        LibOrder.Side indexed side,
        LibOrder.SaleKind indexed saleKind,
        address indexed maker,
        LibOrder.Asset nft,
        Price price,
        uint64 expiry,
        uint64 salt
    );
    // 取消订单
    event LogCancel(OrderKey indexed orderKey, address indexed maker);
    // 订单匹配
    event LogMatch(
        OrderKey indexed makeOrderKey,
        OrderKey indexed takeOrderKey,
        LibOrder.Order makeOrder,
        LibOrder.Order takeOrder,
        uint128 fillPrice
    );
    // ETH 提现事件
    event LogWithdrawETH(address recipient, uint256 amount);
    event BatchMatchInnerError(uint256 offset, bytes msg);
    // 跳过的订单
    event LogSkipOrder(OrderKey orderKey, uint64 salt);
    // 仅允许委托调用
    modifier onlyDelegateCall() {
        _checkDelegateCall();
        _;
    }

    /// @custom:oz-upgrades-unsafe-allow state-variable-immutable state-variable-assignment
    address private immutable self = address(this);

    address private _vault;

    /**
     * @notice Initialize contracts.
     * @param newProtocolShare Default protocol fee.
     * @param newVault easy swap vault address.
     */
    // 部署合约时，会调用 initialize 函数
    function initialize(
        uint128 newProtocolShare, // newProtocolShare - 协议费率
        address newVault, // newVault - 保管库地址
        string memory EIP712Name, // EIP712Name - 合约名称
        string memory EIP712Version // EIP712Version - 版本号
    ) public initializer {
        //使用 initializer 修饰符确保只能初始化一次
        __EasySwapOrderBook_init(
            newProtocolShare,
            newVault,
            EIP712Name,
            EIP712Version
        );
    }

    function __EasySwapOrderBook_init(
        uint128 newProtocolShare,
        address newVault,
        string memory EIP712Name,
        string memory EIP712Version
    ) internal onlyInitializing {
        __EasySwapOrderBook_init_unchained(
            newProtocolShare,
            newVault,
            EIP712Name,
            EIP712Version
        );
    }

    function __EasySwapOrderBook_init_unchained(
        uint128 newProtocolShare,
        address newVault,
        string memory EIP712Name,
        string memory EIP712Version
    ) internal onlyInitializing {
        __Context_init(); // 初始化上下文
        __Ownable_init(_msgSender()); // 初始化所有权
        __ReentrancyGuard_init(); // 初始化重入保护
        __Pausable_init(); // 初始化暂停功能

        __OrderStorage_init(); // 初始化订单存储
        __ProtocolManager_init(newProtocolShare); // 初始化协议管理
        __OrderValidator_init(EIP712Name, EIP712Version); // 初始化订单验证

        setVault(newVault); // 设置保管库地址
    }

    /**
     * @notice Create multiple orders and transfer related assets.
     * @dev If Side=List, you need to authorize the EasySwapVault contract first (creating a List order will transfer the NFT to the order pool).
     * @dev If Side=Bid, you need to pass {value}: the price of the bid (similarly, creating a Bid order will transfer ETH to the order pool).
     * @dev order.maker needs to be msg.sender.
     * @dev order.price cannot be 0.
     * @dev order.expiry needs to be greater than block.timestamp, or 0.
     * @dev order.salt cannot be 0.
     * @param newOrders Multiple order structure data.
     * @return newOrderKeys The unique id of the order is returned in order, if the id is empty, the corresponding order was not created correctly.
     */
    // 批量处理订单，根据订单类型处理资产（NFT 或 ETH）的转移，并返回每个订单的唯一标识符 OrderKey
    // 支持两种订单类型：
    // List：挂单出售 NFT。用户调用 makeOrders 创建 List 类型订单，将 NFT 转移到保管库。
    // Bid：出价购买 NFT。用户调用 makeOrders 创建 Bid 类型订单，将 ETH 转移到保管库。
    // LibOrder.Order[] 订单的数组，每个订单结构包括订单类型、价格、NFT 信息
    function makeOrders(
        LibOrder.Order[] calldata newOrders
    )
        external
        payable
        override
        whenNotPaused
        nonReentrant
        returns (OrderKey[] memory newOrderKeys)
    {
        // 订单的数量
        uint256 orderAmount = newOrders.length;
        // 每个订单的唯一标识符
        newOrderKeys = new OrderKey[](orderAmount);
        // 记录所有 Bid 类型订单的总 ETH 金额
        uint128 ETHAmount; // total eth amount
        for (uint256 i = 0; i < orderAmount; ++i) {
            // Bid 订单的价格
            uint128 buyPrice; // the price of bid order
            if (newOrders[i].side == LibOrder.Side.Bid) {
                // 计算 Bid 类型订单的总价格= 单个 NFT 的价格 × NFT 数
                // Price.unwrap从 Price 类型中提取实际的价格值
                buyPrice =
                    Price.unwrap(newOrders[i].price) *
                    newOrders[i].nft.amount;
            }
            // 验证订单的有效性
            OrderKey newOrderKey = _makeOrderTry(newOrders[i], buyPrice);
            newOrderKeys[i] = newOrderKey;
            if (
                // if the order is not created successfully, the eth will be returned
                OrderKey.unwrap(newOrderKey) !=
                OrderKey.unwrap(LibOrder.ORDERKEY_SENTINEL)
            ) {
                ETHAmount += buyPrice;
            }
        }

        // 用户发送的 ETH 超过了 Bid 类型订单的总金额，退还多余的 ETH
        if (msg.value > ETHAmount) {
            // return the remaining eth，if the eth is not enough, the transaction will be reverted
            _msgSender().safeTransferETH(msg.value - ETHAmount);
        }
    }

    /**
     * @dev Cancels multiple orders by their order keys.
     * @param orderKeys The array of order keys to cancel.
     */
    // 批量取消订单
    function cancelOrders(
        OrderKey[] calldata orderKeys //需要取消的订单的唯一标识符数组
    )
        external
        override
        whenNotPaused //合约未暂停时才能调用
        nonReentrant // 防止重入攻击
        returns (
            bool[] memory successes //返回每个订单的取消结果
        )
    {
        // 初始化返回值数组
        successes = new bool[](orderKeys.length);
        // 遍历订单数组
        for (uint256 i = 0; i < orderKeys.length; ++i) {
            // 尝试取消每个订单
            bool success = _cancelOrderTry(orderKeys[i]);
            // 取消结果存储到 successes 数组中
            successes[i] = success;
        }
    }

    /**
     * @notice Cancels multiple orders by their order keys.
     * @dev newOrder's saleKind, side, maker, and nft must match the corresponding order of oldOrderKey, otherwise it will be skipped; only the price can be modified.
     * @dev newOrder's expiry and salt can be regenerated.
     * @param editDetails The edit details of oldOrderKey and new order info
     * @return newOrderKeys The unique id of the order is returned in order, if the id is empty, the corresponding order was not edit correctly.
     */
    // 批量编辑订单的方法,用户可以通过提供旧订单的唯一标识符（oldOrderKey）和新订单信息（newOrder），修改订单的价格和数量。其他参数（如 saleKind、side、maker 等）必须与旧订单保持一致，否则订单会被跳过
    function editOrders(
        LibOrder.EditDetail[] calldata editDetails //包含多个订单编辑详情的数组
    )
        external
        payable
        override
        whenNotPaused
        nonReentrant
        returns (OrderKey[] memory newOrderKeys)
    {
        // 初始化返回值数组
        newOrderKeys = new OrderKey[](editDetails.length);
        // 遍历订单编辑详情
        uint256 bidETHAmount;
        for (uint256 i = 0; i < editDetails.length; ++i) {
            // 尝试编辑每个订单
            (OrderKey newOrderKey, uint256 bidPrice) = _editOrderTry(
                editDetails[i].oldOrderKey,
                editDetails[i].newOrder
            );
            // 累加 Bid 类型订单的新增 ETH 金额到 bidETHAmount
            bidETHAmount += bidPrice;
            // 将新订单的唯一标识符存储到 newOrderKeys 数组中
            newOrderKeys[i] = newOrderKey;
        }

        if (msg.value > bidETHAmount) {
            // 果用户发送的 ETH 超过了 Bid 类型订单的新增金额，退还多余的 ETH
            _msgSender().safeTransferETH(msg.value - bidETHAmount);
        }
    }

    // 匹配单个买卖订单
    function matchOrder(
        LibOrder.Order calldata sellOrder, //卖单的详细信息，包括订单类型、价格、NFT 信息等
        LibOrder.Order calldata buyOrder //买单的详细信息，包括订单类型、价格、NFT 信息等
    ) external payable override whenNotPaused nonReentrant {
        // 执行订单匹配
        uint256 costValue = _matchOrder(sellOrder, buyOrder, msg.value);
        if (msg.value > costValue) {
            // 如果调用者发送的 ETH 超过了匹配所需的金额，将多余的 ETH 退还给调用者
            _msgSender().safeTransferETH(msg.value - costValue);
        }
    }

    /**
     * @dev Matches multiple orders atomically.
     * @dev If buying NFT, use the "valid sellOrder order" and construct a matching buyOrder order for order matching:
     * @dev    buyOrder.side = Bid, buyOrder.saleKind = FixedPriceForItem, buyOrder.maker = msg.sender,
     * @dev    nft and price values are the same as sellOrder, buyOrder.expiry > block.timestamp, buyOrder.salt != 0;
     * @dev If selling NFT, use the "valid buyOrder order" and construct a matching sellOrder order for order matching:
     * @dev    sellOrder.side = List, sellOrder.saleKind = FixedPriceForItem, sellOrder.maker = msg.sender,
     * @dev    nft and price values are the same as buyOrder, sellOrder.expiry > block.timestamp, sellOrder.salt != 0;
     * @param matchDetails Array of `MatchDetail` structs containing the details of sell and buy order to be matched.
     */
    /// @custom:oz-upgrades-unsafe-allow delegatecall
    // 批量匹配订单,允许用户通过提供多个买卖订单的匹配详情（MatchDetail），一次性完成多个订单的匹配。
    // 该方法会验证每个订单的有效性，并根据订单类型（List 或 Bid）处理资产（NFT 或 ETH）的转移。
    // 如果调用者发送的 ETH 超过了匹配所需的金额，会将多余的 ETH 退还给调用者。
    function matchOrders(
        LibOrder.MatchDetail[] calldata matchDetails //包含多个订单匹配详情的数组
    )
        external
        payable
        override
        whenNotPaused
        nonReentrant
        returns (bool[] memory successes)
    {
        // 初始化返回值数组
        successes = new bool[](matchDetails.length);
        // 初始化变量 buyETHAmount，用于记录买单中已花费的 ETH 总额
        uint128 buyETHAmount;
        // 遍历订单匹配详情
        for (uint256 i = 0; i < matchDetails.length; ++i) {
            LibOrder.MatchDetail calldata matchDetail = matchDetails[i];
            // 使用 delegatecall 调用内部方法 matchOrderWithoutPayback，执行订单匹配逻辑
            // 将 sellOrder、buyOrder 和剩余的 ETH 作为参数传递
            (bool success, bytes memory data) = address(this).delegatecall(
                abi.encodeWithSignature(
                    "matchOrderWithoutPayback((uint8,uint8,address,(uint256,address,uint96),uint128,uint64,uint64),(uint8,uint8,address,(uint256,address,uint96),uint128,uint64,uint64),uint256)",
                    matchDetail.sellOrder,
                    matchDetail.buyOrder,
                    msg.value - buyETHAmount
                )
            );

            if (success) {
                // 如果匹配成功,将结果存储到 successes 数组中
                successes[i] = success;
                // 如果调用者是买单的创建者，解码返回值 data，获取买单的价格，并累加到 buyETHAmount
                if (matchDetail.buyOrder.maker == _msgSender()) {
                    // buy order
                    uint128 buyPrice;
                    buyPrice = abi.decode(data, (uint128));
                    // Calculate ETH the buyer has spent
                    buyETHAmount += buyPrice;
                }
            } else {
                // 匹配失败：触发 BatchMatchInnerError 事件，记录失败的订单索引和错误信息
                emit BatchMatchInnerError(i, data);
            }
        }
        //如果调用者发送的 ETH 超过了匹配所需的金额，将多余的 ETH 退还给调用者
        if (msg.value > buyETHAmount) {
            // return the remaining eth
            _msgSender().safeTransferETH(msg.value - buyETHAmount);
        }
    }

    // 执行单个订单的匹配逻辑
    function matchOrderWithoutPayback(
        LibOrder.Order calldata sellOrder,
        LibOrder.Order calldata buyOrder,
        uint256 msgValue
    )
        external
        payable
        whenNotPaused
        onlyDelegateCall
        returns (
            uint128 costValue //返回匹配订单的成本值
        )
    {
        // 单个订单的匹配
        costValue = _matchOrder(sellOrder, buyOrder, msgValue);
    }

    // 创建订单
    function _makeOrderTry(
        LibOrder.Order calldata order,
        uint128 ETHAmount
    ) internal returns (OrderKey newOrderKey) {
        if (
            order.maker == _msgSender() && // only maker can make order
            Price.unwrap(order.price) != 0 && // price cannot be zero
            order.salt != 0 && // salt cannot be zero
            (order.expiry > block.timestamp || order.expiry == 0) && // expiry must be greater than current block timestamp or no expiry
            filledAmount[LibOrder.hash(order)] == 0 // order cannot be canceled or filled
        ) {
            newOrderKey = LibOrder.hash(order);
            // 根据订单类型（List 或 Bid）处理资产转移
            // deposit asset to vault
            if (order.side == LibOrder.Side.List) {
                // 如果订单类型是 List，将 NFT 转移到保管库
                if (order.nft.amount != 1) {
                    // limit list order amount to 1
                    return LibOrder.ORDERKEY_SENTINEL;
                }
                IEasySwapVault(_vault).depositNFT(
                    newOrderKey,
                    order.maker,
                    order.nft.collection,
                    order.nft.tokenId
                );
            } else if (order.side == LibOrder.Side.Bid) {
                // 如果订单类型是 Bid，将 ETH 转移到保管库
                if (order.nft.amount == 0) {
                    return LibOrder.ORDERKEY_SENTINEL;
                }
                IEasySwapVault(_vault).depositETH{value: uint256(ETHAmount)}(
                    newOrderKey,
                    ETHAmount
                );
            }

            _addOrder(order);
            // 存储订单并触发 LogMake 事件
            emit LogMake(
                newOrderKey,
                order.side,
                order.saleKind,
                order.maker,
                order.nft,
                order.price,
                order.expiry,
                order.salt
            );
        } else {
            emit LogSkipOrder(LibOrder.hash(order), order.salt);
        }
    }

    // 尝试取消指定的订单
    function _cancelOrderTry(
        OrderKey orderKey
    ) internal returns (bool success) {
        // 获取订单信息
        LibOrder.Order memory order = orders[orderKey].order;
        //验证调用者是否为订单的创建者.验证订单是否未完全成交
        if (
            order.maker == _msgSender() &&
            filledAmount[orderKey] < order.nft.amount // only unfilled order can be canceled
        ) {
            // 计算订单的哈希值。
            OrderKey orderHash = LibOrder.hash(order);
            // 将订单从存储中移除
            _removeOrder(order);
            // withdraw asset from vault
            if (order.side == LibOrder.Side.List) {
                // 如果订单类型是 List（挂单出售 NFT）
                // 调用保管库合约的 withdrawNFT 方法，将 NFT 返还给订单的创建者
                IEasySwapVault(_vault).withdrawNFT(
                    orderHash,
                    order.maker,
                    order.nft.collection,
                    order.nft.tokenId
                );
            } else if (order.side == LibOrder.Side.Bid) {
                // 如果订单类型是 Bid（出价购买 NFT）
                // 计算未成交的 NFT 数量
                uint256 availNFTAmount = order.nft.amount -
                    filledAmount[orderKey];
                // 调用保管库合约的 withdrawETH 方法，将未使用的 ETH 返还给订单的创建者
                IEasySwapVault(_vault).withdrawETH(
                    orderHash,
                    Price.unwrap(order.price) * availNFTAmount, // the withdraw amount of eth
                    order.maker
                );
            }
            // 调用 _cancelOrder 更新订单状态
            _cancelOrder(orderKey);
            success = true;
            // 触发 LogCancel 事件，记录订单取消信息
            emit LogCancel(orderKey, order.maker);
        } else {
            // 验证失败，触发 LogSkipOrder 事件并退出
            emit LogSkipOrder(orderKey, order.salt);
        }
    }

    // 尝试编辑订单
    function _editOrderTry(
        OrderKey oldOrderKey, //旧订单的唯一标识符
        LibOrder.Order calldata newOrder //新订单的详细信息
    ) internal returns (OrderKey newOrderKey, uint256 deltaBidPrice) {
        //newOrderKey:新订单的唯一标识符; deltaBidPrice: Bid 类型订单的新增 ETH 金额
        LibOrder.Order memory oldOrder = orders[oldOrderKey].order;

        // check order, only the price and amount can be modified
        // 验证新订单的 saleKind、side、maker、nft 等参数是否与旧订单一致
        if (
            (oldOrder.saleKind != newOrder.saleKind) ||
            (oldOrder.side != newOrder.side) ||
            (oldOrder.maker != newOrder.maker) ||
            (oldOrder.nft.collection != newOrder.nft.collection) ||
            (oldOrder.nft.tokenId != newOrder.nft.tokenId) ||
            filledAmount[oldOrderKey] >= oldOrder.nft.amount // order cannot be canceled or filled验证旧订单是否未完全成交
        ) {
            //如果验证失败，触发 LogSkipOrder 事件并返回
            emit LogSkipOrder(oldOrderKey, oldOrder.salt);
            return (LibOrder.ORDERKEY_SENTINEL, 0);
        }

        // check new order is valid
        // 验证新订单的参数
        if (
            // 创建者是否为调用者,新订单的 salt 是否为非零,过期时间是否有效,是否未被取消或成交
            newOrder.maker != _msgSender() ||
            newOrder.salt == 0 ||
            (newOrder.expiry < block.timestamp && newOrder.expiry != 0) ||
            filledAmount[LibOrder.hash(newOrder)] != 0 // order cannot be canceled or filled
        ) {
            emit LogSkipOrder(oldOrderKey, newOrder.salt);
            return (LibOrder.ORDERKEY_SENTINEL, 0);
        }

        // 取消旧订单
        uint256 oldFilledAmount = filledAmount[oldOrderKey];
        // 调用 _removeOrder 和 _cancelOrder，从存储中移除旧订单
        _removeOrder(oldOrder); // remove order from order storage
        _cancelOrder(oldOrderKey); // cancel order from order book
        emit LogCancel(oldOrderKey, oldOrder.maker); //记录旧订单的取消信息
        // 将新订单添加到存储中
        newOrderKey = _addOrder(newOrder); // add new order to order storage

        // make new order
        // 根据订单类型（List 或 Bid），更新相关资产
        if (oldOrder.side == LibOrder.Side.List) {
            // 订单类型是 List（挂单出售 NFT）,调用保管库合约的 editNFT 方法，更新 NFT 的存储信息
            IEasySwapVault(_vault).editNFT(oldOrderKey, newOrderKey);
        } else if (oldOrder.side == LibOrder.Side.Bid) {
            // 订单类型是 Bid（出价购买 NFT）
            // 计算旧订单和新订单的剩余价格
            uint256 oldRemainingPrice = Price.unwrap(oldOrder.price) *
                (oldOrder.nft.amount - oldFilledAmount);
            uint256 newRemainingPrice = Price.unwrap(newOrder.price) *
                newOrder.nft.amount;
            // 如果新订单的价格高于旧订单，计算新增的 ETH 金额并调用 editETH 方法更新保管库中的 ETH
            if (newRemainingPrice > oldRemainingPrice) {
                deltaBidPrice = newRemainingPrice - oldRemainingPrice;
                IEasySwapVault(_vault).editETH{value: uint256(deltaBidPrice)}(
                    oldOrderKey,
                    newOrderKey,
                    oldRemainingPrice,
                    newRemainingPrice,
                    oldOrder.maker
                );
            } else {
                //如果新订单的价格低于或等于旧订单，直接调用 editETH 方法更新保管库中的 ETH。
                IEasySwapVault(_vault).editETH(
                    oldOrderKey,
                    newOrderKey,
                    oldRemainingPrice,
                    newRemainingPrice,
                    oldOrder.maker
                );
            }
        }
        //记录新订单的创建信息
        emit LogMake(
            newOrderKey,
            newOrder.side,
            newOrder.saleKind,
            newOrder.maker,
            newOrder.nft,
            newOrder.price,
            newOrder.expiry,
            newOrder.salt
        );
    }

    // 匹配单个买卖订单。它会验证订单的有效性和匹配条件，并根据调用者的身份（卖单创建者或买单创建者）处理资产（NFT 或 ETH）的转移。
    // 匹配成功后，触发 LogMatch 事件记录匹配信息
    function _matchOrder(
        LibOrder.Order calldata sellOrder, //卖单的详细信息
        LibOrder.Order calldata buyOrder, //买单的详细信息
        uint256 msgValue //调用者发送的 ETH 金额
    ) internal returns (uint128 costValue) {
        //costValue 匹配订单的成本值（即买单的价格）
        //计算订单的唯一标识符
        OrderKey sellOrderKey = LibOrder.hash(sellOrder);
        OrderKey buyOrderKey = LibOrder.hash(buyOrder);
        // 验证订单的匹配条件
        _isMatchAvailable(sellOrder, buyOrder, sellOrderKey, buyOrderKey);
        //如果调用者是卖单的创建者
        if (_msgSender() == sellOrder.maker) {
            // sell order
            // accept bid
            require(msgValue == 0, "HD: value > 0"); // 卖单不能接受 ETH
            bool isSellExist = orders[sellOrderKey].order.maker != address(0); // 检查卖单是否存在// check if sellOrder exist in order storage
            _validateOrder(sellOrder, isSellExist);
            _validateOrder(orders[buyOrderKey].order, false); // 验证买单是否有效// check if exist in order storage

            uint128 fillPrice = Price.unwrap(buyOrder.price); // 获取买单的价格// the price of bid order
            if (isSellExist) {
                // check if sellOrder exist in order storage , del&fill if exist
                _removeOrder(sellOrder); // 从存储中移除卖单
                _updateFilledAmount(sellOrder.nft.amount, sellOrderKey); // 更新卖单的成交状态// sell order totally filled
            }
            // 更新买单的成交状态
            _updateFilledAmount(filledAmount[buyOrderKey] + 1, buyOrderKey);
            emit LogMatch(
                sellOrderKey,
                buyOrderKey,
                sellOrder,
                buyOrder,
                fillPrice
            );

            // transfer nft&eth
            // 转移资产（卖单逻辑）从保管库中提取买单的 ETH
            IEasySwapVault(_vault).withdrawETH(
                buyOrderKey,
                fillPrice,
                address(this)
            );
            // 计算协议费用并将剩余的 ETH 转移给卖单创建者
            uint128 protocolFee = _shareToAmount(fillPrice, protocolShare);
            sellOrder.maker.safeTransferETH(fillPrice - protocolFee);
            //如果卖单存在于存储中，从保管库中提取 NFT；否则直接从卖单创建者转移 NFT 给买单创建者。
            if (isSellExist) {
                IEasySwapVault(_vault).withdrawNFT(
                    sellOrderKey,
                    buyOrder.maker,
                    sellOrder.nft.collection,
                    sellOrder.nft.tokenId
                );
            } else {
                IEasySwapVault(_vault).transferERC721(
                    sellOrder.maker,
                    buyOrder.maker,
                    sellOrder.nft
                );
            }
        } else if (_msgSender() == buyOrder.maker) {
            // 如果调用者是买单的创建者
            // buy order
            // accept list
            bool isBuyExist = orders[buyOrderKey].order.maker != address(0); // 检查买单是否存在
            //验证买单和卖单的有效性
            _validateOrder(orders[sellOrderKey].order, false); // 验证卖单是否有效// check if exist in order storage
            _validateOrder(buyOrder, isBuyExist);

            uint128 buyPrice = Price.unwrap(buyOrder.price);
            uint128 fillPrice = Price.unwrap(sellOrder.price);
            if (!isBuyExist) {
                //如果买单不存在，确保调用者发送的 ETH 足够支付卖单的价格。
                require(msgValue >= fillPrice, "HD: value < fill price");
            } else {
                // 如果买单存在，从保管库中提取买单的 ETH，并移除买单。
                require(buyPrice >= fillPrice, "HD: buy price < fill price");
                IEasySwapVault(_vault).withdrawETH(
                    buyOrderKey,
                    buyPrice,
                    address(this)
                );
                // check if buyOrder exist in order storage , del&fill if exist
                _removeOrder(buyOrder); // 从存储中移除买单
                _updateFilledAmount(filledAmount[buyOrderKey] + 1, buyOrderKey);
            }
            _updateFilledAmount(sellOrder.nft.amount, sellOrderKey);

            emit LogMatch(
                buyOrderKey,
                sellOrderKey,
                buyOrder,
                sellOrder,
                fillPrice
            );

            // transfer nft&eth
            // 转移资产（买单逻辑）
            // 计算协议费用并将剩余的 ETH 转移给卖单创建者。
            uint128 protocolFee = _shareToAmount(fillPrice, protocolShare);
            sellOrder.maker.safeTransferETH(fillPrice - protocolFee);
            if (buyPrice > fillPrice) {
                // 如果买单的价格高于卖单的价格，将多余的 ETH 退还给买单创建者。
                buyOrder.maker.safeTransferETH(buyPrice - fillPrice);
            }
            //从保管库中提取 NFT 给买单创建者
            IEasySwapVault(_vault).withdrawNFT(
                sellOrderKey,
                buyOrder.maker,
                sellOrder.nft.collection,
                sellOrder.nft.tokenId
            );
            costValue = isBuyExist ? 0 : buyPrice;
        } else {
            // 如果调用者既不是卖单创建者也不是买单创建者
            revert("HD: sender invalid");
        }
    }

    // 验证买卖订单是否满足匹配条件。它会检查订单的类型、资产一致性、订单状态等条件。
    // 如果任意条件不满足，方法会通过 require 抛出异常，阻止匹配操作。
    function _isMatchAvailable(
        LibOrder.Order memory sellOrder, //卖单的详细信息
        LibOrder.Order memory buyOrder, //买单的详细信息
        OrderKey sellOrderKey, //卖单的唯一标识符
        OrderKey buyOrderKey //买单的唯一标识符
    ) internal view {
        // 验证卖单和买单是否为同一个订单
        require(
            OrderKey.unwrap(sellOrderKey) != OrderKey.unwrap(buyOrderKey),
            "HD: same order"
        );
        //证卖单的类型是否为 List，买单的类型是否为 Bid
        require(
            sellOrder.side == LibOrder.Side.List &&
                buyOrder.side == LibOrder.Side.Bid,
            "HD: side mismatch"
        );
        //验证卖单的销售类型是否为 FixedPriceForItem
        require(
            sellOrder.saleKind == LibOrder.SaleKind.FixedPriceForItem,
            "HD: kind mismatch"
        );
        // 验证卖单和买单的创建者是否为同一地址
        require(sellOrder.maker != buyOrder.maker, "HD: same maker");
        //验证买卖订单的资产是否一致
        // 如果买单的销售类型为 FixedPriceForCollection，允许资产集合匹配
        // 否则，验证卖单和买单的 collection 和 tokenId 是否相同
        require( // check if the asset is the same
            buyOrder.saleKind == LibOrder.SaleKind.FixedPriceForCollection ||
                (sellOrder.nft.collection == buyOrder.nft.collection &&
                    sellOrder.nft.tokenId == buyOrder.nft.tokenId),
            "HD: asset mismatch"
        );
        // 验证卖单和买单的成交数量是否小于订单的总数量
        require(
            filledAmount[sellOrderKey] < sellOrder.nft.amount &&
                filledAmount[buyOrderKey] < buyOrder.nft.amount,
            "HD: order closed"
        );
    }

    /**
     * @notice caculate amount based on share.
     * @param total the total amount.
     * @param share the share in base point.
     */
    // 根据比例（share）计算总金额（total）中对应的部分金额。计算协议费用或分成金额
    function _shareToAmount(
        uint128 total, //总金额
        uint128 share //比例
    ) internal pure returns (uint128) {
        // 计算总金额与比例的乘积,将乘积除以 LibPayInfo.TOTAL_SHARE，将比例基点转换为实际比例
        return (total * share) / LibPayInfo.TOTAL_SHARE;
    }

    function _checkDelegateCall() private view {
        require(address(this) != self);
    }

    // 允许合约所有者设置或更新保管库地址的方法
    function setVault(address newVault) public onlyOwner {
        require(newVault != address(0), "HD: zero address");
        _vault = newVault;
    }

    // 允许合约所有者提取合约中 ETH
    function withdrawETH(
        address recipient, //接收 ETH 的地址
        uint256 amount //要提取的 ETH 金额
    ) external nonReentrant onlyOwner {
        // 将指定金额的 ETH 转移到接收地址
        recipient.safeTransferETH(amount);
        // 记录提取操作的接收地址和金额
        emit LogWithdrawETH(recipient, amount);
    }

    function pause() external onlyOwner {
        _pause();
    }

    function unpause() external onlyOwner {
        _unpause();
    }

    receive() external payable {}

    uint256[50] private __gap;
}
