//"SPDX-License-Identifier: UNLICENSED"
pragma solidity 0.6.12;

// 引入接口和库
import "../interfaces/IAdmin.sol";
import "../interfaces/ISalesFactory.sol";
import "../interfaces/IAllocationStaking.sol";
import "../interfaces/IERC20Metadata.sol";
import "@openzeppelin/contracts/cryptography/ECDSA.sol";
import "@openzeppelin/contracts/token/ERC20/SafeERC20.sol";
import "@openzeppelin/contracts/utils/ReentrancyGuard.sol";

// C2N 代币销售合约
// 支持注册、参与购买、线性释放等功能
contract C2NSale is ReentrancyGuard {
    using ECDSA for bytes32;
    using SafeMath for uint256;
    using SafeERC20 for IERC20;

    // Pointer to Allocation staking contract
    // AllocationStaking合约接口
    IAllocationStaking public allocationStakingContract;
    // Pointer to sales factory contract0
    // SalesFactory合约接口
    ISalesFactory public factory;
    // Admin contract
    // Admin合约接口
    IAdmin public admin;

    // 销售信息结构体
    struct Sale {
        // Token being sold
        // 正在销售的代币
        IERC20 token;
        // Is sale created
        // 销售是否已创建
        bool isCreated;
        // Are earnings withdrawn
        // 收益是否已提取
        bool earningsWithdrawn;
        // Is leftover withdrawn
        // 剩余代币是否已提取
        bool leftoverWithdrawn;
        // Have tokens been deposited
        // 代币是否已存入
        bool tokensDeposited;
        // Address of sale owner
        // 销售所有者地址
        address saleOwner;
        // Price of the token quoted in ETH
        // 代币价格（以 ETH 计价）
        uint256 tokenPriceInETH;
        // Amount of tokens to sell
        // 待销售代币数量
        uint256 amountOfTokensToSell;
        // Total tokens being sold
        // 总的已经销售的代币数量
        uint256 totalTokensSold;
        // Total ETH Raised
        // 总筹集的 ETH 数量
        uint256 totalETHRaised;
        // Sale start time
        // 销售开始时间
        uint256 saleStart;
        // Sale end time
        // 销售结束时间（项目方只能在销售结束后才能提取合约中收获的ETH和剩余代币）
        uint256 saleEnd;
        // When tokens can be withdrawn
        // 代币可提取的时间（购买者可以在该时间后提取其购买的代币,同时要等待线性解锁时间）
        uint256 tokensUnlockTime;
        // maxParticipation
        // 最大购买金额（单次）
        uint256 maxParticipation;
    }

    // Participation structure
    // 购买者参与购买的信息结构体
    struct Participation {
        uint256 amountBought;           // 购买的代币数量
        uint256 amountETHPaid;          // 支付的 ETH 数量
        uint256 timeParticipated;       // 参与时间
        bool[] isPortionWithdrawn;      // 各部分是否已提取（线性解锁涉及）
    }

    // 注册信息结构体（这是项目方注册参与销售的时间窗口）
    struct Registration {
        uint256 registrationTimeStarts; // 注册开始时间
        uint256 registrationTimeEnds;   // 注册结束时间
        uint256 numberOfRegistrants;    // 注册的参与购买用户数量
    }

    // Sale
    // 销售信息
    Sale public sale;
    // Registration
    // 注册信息
    Registration public registration;
    // Number of users participated in the sale.
    // 参与购买的用户数量
    uint256 public numberOfParticipants;
    // Mapping user to his participation
    // 购买者地址 => 购买者参与购买的信息
    mapping(address => Participation) public userToParticipation;
    // Mapping if user is registered or not
    // 购买者是否已注册（参与购买）
    mapping(address => bool) public isRegistered;
    // mapping if user is participated or not
    // 购买者是否已参与购买
    mapping(address => bool) public isParticipated;
    // Times when portions are getting unlocked
    // 各部分代币解锁时间
    uint256[] public vestingPortionsUnlockTime;
    // Percent of the participation user can withdraw
    // 各部分可提取的份额
    uint256[] public vestingPercentPerPortion;
    //Precision for percent for portion vesting
    // 可释放的总份额
    uint256 public portionVestingPrecision;
    // Max vesting time shift
    // 可延长的最大代币解锁时间偏移量（秒）
    uint256 public maxVestingTimeShift;

    // Restricting calls only to sale owner
    // 修饰符：仅销售所有者（项目方）可调用
    modifier onlySaleOwner() {
        require(msg.sender == sale.saleOwner, "OnlySaleOwner:: Restricted");
        _;
    }

    // 修饰符：仅管理员可调用
    modifier onlyAdmin() {
        require(
            admin.isAdmin(msg.sender),
            "Only admin can call this function."
        );
        _;
    }

    // 事件定义
    event TokensSold(address user, uint256 amount);                              // 代币售出事件
    event UserRegistered(address user);                                          // 购买者注册（参与购买）事件（购买者地址）
    event TokenPriceSet(uint256 newPrice);                                       // 代币价格设置事件
    event MaxParticipationSet(uint256 maxParticipation);                         // 最大募集金额设置事件
    event TokensWithdrawn(address user, uint256 amount);                         // 代币提取事件
    event SaleCreated(
        address saleOwner,
        uint256 tokenPriceInETH,
        uint256 amountOfTokensToSell,
        uint256 saleEnd
    );                                                                           // 销售创建事件
    event StartTimeSet(uint256 startTime);                                       // 开始时间设置事件
    event RegistrationTimeSet(
        uint256 registrationTimeStarts,
        uint256 registrationTimeEnds
    );                                                                           // 注册时间设置事件

    // Constructor, always initialized through SalesFactory
    // 构造函数：总是通过 SalesFactory 初始化
    // _admin: Admin合约地址
    // _allocationStaking: AllocationStaking合约地址
    constructor(address _admin, address _allocationStaking) public {
        require(_admin != address(0));
        require(_allocationStaking != address(0));
        admin = IAdmin(_admin);
        factory = ISalesFactory(msg.sender);
        allocationStakingContract = IAllocationStaking(_allocationStaking);
    }

    /// @notice         Function to set vesting params
    /// @notice         设置线性释放参数函数
    /// @param          _unlockingTimes 各部分代币的解锁时间点数组
    /// @param          _percents 对应的释放百分比数组
    /// @param          _maxVestingTimeShift 最大可偏移的时间量（秒）
    function setVestingParams(
        uint256[] memory _unlockingTimes,
        uint256[] memory _percents,
        uint256 _maxVestingTimeShift
    ) external onlyAdmin {
        require(
            vestingPercentPerPortion.length == 0 &&
            vestingPortionsUnlockTime.length == 0
        );
        require(_unlockingTimes.length == _percents.length);
        require(portionVestingPrecision > 0, "Safeguard for making sure setSaleParams get first called.");
        require(_maxVestingTimeShift <= 30 days, "Maximal shift is 30 days.");

        // Set max vesting time shift
        maxVestingTimeShift = _maxVestingTimeShift;

        uint256 sum;

        for (uint256 i = 0; i < _unlockingTimes.length; i++) {
            vestingPortionsUnlockTime.push(_unlockingTimes[i]);
            vestingPercentPerPortion.push(_percents[i]);
            sum += _percents[i];
        }

        require(sum == portionVestingPrecision, "Percent distribution issue.");
    }

    /// @notice         延长各部分代币解锁时间（仅限管理员，只能调用一次）
    /// @param          timeToShift 需要延后的时间量（秒）
    function shiftVestingUnlockingTimes(uint256 timeToShift)
    external
    onlyAdmin
    {
        require(
            timeToShift > 0 && timeToShift < maxVestingTimeShift,
            "Shift must be nonzero and smaller than maxVestingTimeShift."
        );

        // Time can be shifted only once.
        maxVestingTimeShift = 0;

        for (uint256 i = 0; i < vestingPortionsUnlockTime.length; i++) {
            vestingPortionsUnlockTime[i] = vestingPortionsUnlockTime[i].add(
                timeToShift
            );
        }
    }

    /// @notice     Admin function to set sale parameters
    /// @notice     管理员设置销售参数函数
    /// @param      _token 销售代币合约地址
    /// @param      _saleOwner 销售所有者地址
    /// @param      _tokenPriceInETH 代币价格（wei单位）
    /// @param      _amountOfTokensToSell 销售代币总量
    /// @param      _saleEnd 销售结束时间戳
    /// @param      _tokensUnlockTime 代币可提取时间戳
    /// @param      _portionVestingPrecision 可释放的总份额
    /// @param      _maxParticipation 单个用户最大参与金额（wei）
    function setSaleParams(
        address _token,
        address _saleOwner,
        uint256 _tokenPriceInETH,
        uint256 _amountOfTokensToSell,
        uint256 _saleEnd,
        uint256 _tokensUnlockTime,
        uint256 _portionVestingPrecision,
        uint256 _maxParticipation
    ) external onlyAdmin {
        require(!sale.isCreated, "setSaleParams: Sale is already created.");
        require(
            _saleOwner != address(0),
            "setSaleParams: Sale owner address can not be 0."
        );
        require(
            _tokenPriceInETH != 0 &&
            _amountOfTokensToSell != 0 &&
            _saleEnd > block.timestamp &&
            _tokensUnlockTime > block.timestamp &&
            _maxParticipation > 0,
            "setSaleParams: Bad input"
        );
        require(_portionVestingPrecision >= 100, "Should be at least 100");

        // Set params
        sale.token = IERC20(_token);
        sale.isCreated = true;
        sale.saleOwner = _saleOwner;
        sale.tokenPriceInETH = _tokenPriceInETH;
        sale.amountOfTokensToSell = _amountOfTokensToSell;
        sale.saleEnd = _saleEnd;
        sale.tokensUnlockTime = _tokensUnlockTime;
        sale.maxParticipation = _maxParticipation;

        // Set portion vesting precision
        portionVestingPrecision = _portionVestingPrecision;
        // Emit event
        emit SaleCreated(
            sale.saleOwner,
            sale.tokenPriceInETH,
            sale.amountOfTokensToSell,
            sale.saleEnd
        );
    }

    // @notice     Function to retroactively set sale token address, can be called only once,
    //             after initial contract creation has passed. Added as an options for teams which
    //             are not having token at the moment of sale launch.
    // @notice     事后设置销售代币地址函数（仅可调用一次），用于销售启动时代币尚未就绪的情况
    function setSaleToken(
        address saleToken
    )
    external
    onlyAdmin
    {
        require(address(sale.token) == address(0));
        sale.token = IERC20(saleToken);
    }


    /// @notice     Function to set registration period parameters
    /// @notice     设置项目方注册时间参数函数（管理员调用）
    /// @param      _registrationTimeStarts 注册开始时间戳
    /// @param      _registrationTimeEnds 注册结束时间戳
    function setRegistrationTime(
        uint256 _registrationTimeStarts,
        uint256 _registrationTimeEnds
    ) external onlyAdmin {
        require(sale.isCreated);
        require(registration.registrationTimeStarts == 0);
        require(
            _registrationTimeStarts >= block.timestamp &&
            _registrationTimeEnds > _registrationTimeStarts
        );
        require(_registrationTimeEnds < sale.saleEnd);

        // 项目方需要在销售开始前注册，否则无法参与销售
        if (sale.saleStart > 0) {
            require(_registrationTimeEnds < sale.saleStart, "registrationTimeEnds >= sale.saleStart is not allowed");
        }

        registration.registrationTimeStarts = _registrationTimeStarts;
        registration.registrationTimeEnds = _registrationTimeEnds;

        emit RegistrationTimeSet(
            registration.registrationTimeStarts,
            registration.registrationTimeEnds
        );
    }

    // 设置销售开始时间（管理员调用）
    function setSaleStart(
        uint256 starTime
    ) external onlyAdmin {
        require(sale.isCreated, "sale is not created.");
        require(sale.saleStart == 0, "setSaleStart: starTime is set already.");
        require(starTime > registration.registrationTimeEnds, "start time should greater than registrationTimeEnds.");
        require(starTime < sale.saleEnd, "start time should less than saleEnd time");
        require(starTime >= block.timestamp, "start time should be in the future.");
        sale.saleStart = starTime;

        // Fire event
        emit StartTimeSet(sale.saleStart);
    }

    /// @notice     Registration for sale.
    /// @notice     购买者注册参与购买（购买者调用）
    /// @param      signature 后台签名的消息（用于验证用户资格，需要有管理员的签名）
    /// @param      pid 质押池ID（用于锁定用户质押）

    // 合约本身不直接检查质押数量，而是通过链下后台签名验证(signature)来实现"质押越多 → 参与额度越大"的机制。后台可以根据质押数量灵活决定每个用户的购买额度

    // 质押代币（C2N）: 购买者在质押池中质押的已有代币，用于获得参与资格/权重

    // 销售代币（新项目代币）: 购买者用 ETH 购买的新代币

    // 流程: 
    // 1. 购买者已有 C2N 代币
    //         ↓
    // 2. 将 C2N 质押到 AllocationStaking 池子
    //         ↓
    // 3. 注册参与销售（质押被锁定到 saleEnd）
    //         ↓
    // 4. 用 ETH 购买新项目代币
    //         ↓
    // 5. 按释放计划提取新代币
    //         ↓
    // 6. 销售结束后，可以取出质押的 C2N
    function registerForSale(bytes memory signature, uint256 pid)
    external
    {
        require(
            block.timestamp >= registration.registrationTimeStarts &&
            block.timestamp <= registration.registrationTimeEnds,
            "Registration gate is closed."
        );
        require(
            checkRegistrationSignature(signature, msg.sender),
            "Invalid signature"
        );
        require(
            !isRegistered[msg.sender],
            "User can not register twice."
        );
        isRegistered[msg.sender] = true;

        // Lock users stake
        allocationStakingContract.setTokensUnlockTime(
            pid,
            msg.sender,
            sale.saleEnd
        );

        // Increment number of registered users
        registration.numberOfRegistrants++;
        // Emit Registration event
        emit UserRegistered(msg.sender);
    }

    /// @notice     Admin function, to update token price before sale to match the closest $ desired rate.
    /// @dev        This will be updated with an oracle during the sale every N minutes, so the users will always
    ///             pay initialy set $ value of the token. This is to reduce reliance on the ETH volatility.
    /// @notice     管理员更新代币价格函数（使用预言机减少 ETH 价格波动影响）
    function updateTokenPriceInETH(uint256 price) external onlyAdmin {
        require(price > 0, "Price can not be 0.");
        // Allowing oracle to run and change the sale value
        sale.tokenPriceInETH = price;
        emit TokenPriceSet(price);
    }

    /// @notice     Admin function to postpone the sale
    /// @notice     管理员推迟开始销售时间（销售开始前）
    function postponeSale(uint256 timeToShift) external onlyAdmin {
        require(
            block.timestamp < sale.saleStart,
            "sale already started."
        );
        //  postpone registration start time
        sale.saleStart = sale.saleStart.add(timeToShift);
        require(
            sale.saleStart + timeToShift < sale.saleEnd,
            "Start time can not be greater than end time."
        );
    }

    /// @notice     Function to extend registration period
    /// @notice     管理员延长注册结束时间
    function extendRegistrationPeriod(uint256 timeToAdd) external onlyAdmin {
        require(
            registration.registrationTimeEnds.add(timeToAdd) <
            sale.saleStart,
            "Registration period overflows sale start."
        );

        registration.registrationTimeEnds = registration
        .registrationTimeEnds
        .add(timeToAdd);
    }

    /// @notice     Admin function to set max participation before sale start
    /// @notice     管理员设置单次最大购买金额（销售开始前）
    function setCap(uint256 cap)
    external
    onlyAdmin
    {
        require(
            block.timestamp < sale.saleStart,
            "sale already started."
        );

        require(cap > 0, "Can't set max participation to 0");

        sale.maxParticipation = cap;

        emit MaxParticipationSet(sale.maxParticipation);
    }

    // Function for owner to deposit tokens, can be called only once.
    // 项目方存入待销售代币（仅可调用一次）
    function depositTokens() external onlySaleOwner {
        require(
            !sale.tokensDeposited, "Deposit can be done only once"
        );

        sale.tokensDeposited = true;

        sale.token.safeTransferFrom(
            msg.sender,
            address(this),
            sale.amountOfTokensToSell
        );
    }

    // Function to participate in the sales
    // 参与购买（用户调用）
    // @param signature 后台签名的消息（用于验证用户资格，需要有管理员的签名）
    // @param amount 用户可购买的代币数量（单位：代币的单位）
    // @dev 该函数需要支付 ETH 作为购买代币的费用
    function participate(
        bytes memory signature,
        uint256 amount
    ) external payable {

        require(
            amount <= sale.maxParticipation,
            "Overflowing maximal participation for sale."
        );

        // User must have registered for the round in advance
        require(
            isRegistered[msg.sender],
            "Not registered for this sale."
        );

        // Verify the signature
        // 检查用户提供的签名是否有效，用于验证用户资格
        // 签名内容：用户地址、最大购买金额
        require(
            checkParticipationSignature(
                signature,
                msg.sender,
                amount
            ),
            "Invalid signature. Verification failed"
        );

        // Verify the timestamp
        require(
            block.timestamp >= sale.saleStart &&
            block.timestamp < sale.saleEnd, "sale didn't start or it's ended."
        );

        // Check user haven't participated before
        // 检查用户是否已参与购买，防止重复购买
        require(!isParticipated[msg.sender], "User can participate only once.");

        // Disallow contract calls.
        // 禁止合约调用，只能由用户(EOA账户)直接调用
        require(msg.sender == tx.origin, "Only direct contract calls.");

        // Compute the amount of tokens user is buying
        uint256 amountOfTokensBuying =
        (msg.value).mul(uint(10) ** IERC20Metadata(address(sale.token)).decimals()).div(sale.tokenPriceInETH);

        // Must buy more than 0 tokens
        require(amountOfTokensBuying > 0, "Can't buy 0 tokens");

        // Check in terms of user allo
        require(
            amountOfTokensBuying <= amount,
            "Trying to buy more than allowed."
        );

        // Increase amount of sold tokens
        sale.totalTokensSold = sale.totalTokensSold.add(amountOfTokensBuying);

        // Increase amount of ETH raised
        sale.totalETHRaised = sale.totalETHRaised.add(msg.value);

        bool[] memory _isPortionWithdrawn = new bool[](
            vestingPortionsUnlockTime.length
        );

        // Create participation object
        Participation memory p = Participation({
        amountBought : amountOfTokensBuying,
        amountETHPaid : msg.value,
        timeParticipated : block.timestamp,
        isPortionWithdrawn : _isPortionWithdrawn
        });

        // Add participation for user.
        userToParticipation[msg.sender] = p;
        // Mark user is participated
        isParticipated[msg.sender] = true;
        // Increment number of participants in the Sale.
        numberOfParticipants++;

        emit TokensSold(msg.sender, amountOfTokensBuying);
    }

    /// Users can claim their participation
    /// 购买者提取代币函数（购买者调用）
    /// @param      portionId 要提取的部分ID（索引，从0开始）
    function withdrawTokens(uint256 portionId) external {
        require(
            block.timestamp >= sale.tokensUnlockTime,
            "Tokens can not be withdrawn yet."
        );
        require(
            portionId < vestingPercentPerPortion.length,
            "Portion id out of range."
        );

        Participation storage p = userToParticipation[msg.sender];

        if (
            !p.isPortionWithdrawn[portionId] &&
        vestingPortionsUnlockTime[portionId] <= block.timestamp
        ) {
            p.isPortionWithdrawn[portionId] = true;
            uint256 amountWithdrawing = p
            .amountBought
            .mul(vestingPercentPerPortion[portionId])
            .div(portionVestingPrecision);

            // Withdraw percent which is unlocked at that portion
            if (amountWithdrawing > 0) {
                sale.token.safeTransfer(msg.sender, amountWithdrawing);
                emit TokensWithdrawn(msg.sender, amountWithdrawing);
            }
        } else {
            revert("Tokens already withdrawn or portion not unlocked yet.");
        }
    }

    // Expose function where user can withdraw multiple unlocked portions at once.
    // 用户一次提取多个已解锁部分函数
    function withdrawMultiplePortions(uint256 [] calldata portionIds) external {
        uint256 totalToWithdraw = 0;

        Participation storage p = userToParticipation[msg.sender];

        for (uint i = 0; i < portionIds.length; i++) {
            uint256 portionId = portionIds[i];
            require(portionId < vestingPercentPerPortion.length);

            if (
                !p.isPortionWithdrawn[portionId] &&
            vestingPortionsUnlockTime[portionId] <= block.timestamp
            ) {
                p.isPortionWithdrawn[portionId] = true;
                uint256 amountWithdrawing = p
                .amountBought
                .mul(vestingPercentPerPortion[portionId])
                .div(portionVestingPrecision);
                // Withdraw percent which is unlocked at that portion
                totalToWithdraw = totalToWithdraw.add(amountWithdrawing);
            }
        }

        if (totalToWithdraw > 0) {
            sale.token.safeTransfer(msg.sender, totalToWithdraw);
            emit TokensWithdrawn(msg.sender, totalToWithdraw);
        }
    }

    // Internal function to handle safe transfer
    // 内部安全转账 ETH 函数
    function safeTransferETH(address to, uint256 value) internal {
        (bool success,) = to.call{value : value}(new bytes(0));
        require(success);
    }

    /// Function to withdraw all the earnings and the leftover of the sale contract.
    /// 提取收益和剩余代币（项目方调用）
    function withdrawEarningsAndLeftover() external onlySaleOwner {
        withdrawEarningsInternal();
        withdrawLeftoverInternal();
    }

    // Function to withdraw only earnings
    // 仅提取收益函数
    function withdrawEarnings() external onlySaleOwner {
        withdrawEarningsInternal();
    }

    // Function to withdraw only leftover
    // 仅提取剩余代币函数
    function withdrawLeftover() external onlySaleOwner {
        withdrawLeftoverInternal();
    }

    // function to withdraw earnings
    // 内部提取收益函数
    function withdrawEarningsInternal() internal {
        // Make sure sale ended
        require(block.timestamp >= sale.saleEnd, "sale is not ended yet.");

        // Make sure owner can't withdraw twice
        require(!sale.earningsWithdrawn, "owner can't withdraw earnings twice");
        sale.earningsWithdrawn = true;
        // Earnings amount of the owner in ETH
        uint256 totalProfit = sale.totalETHRaised;

        safeTransferETH(msg.sender, totalProfit);
    }

    // Function to withdraw leftover
    // 内部提取剩余代币函数
    function withdrawLeftoverInternal() internal {
        // Make sure sale ended
        require(block.timestamp >= sale.saleEnd, "sale is not ended yet.");

        // Make sure owner can't withdraw twice
        require(!sale.leftoverWithdrawn, "owner can't withdraw leftover twice");
        sale.leftoverWithdrawn = true;

        // Amount of tokens which are not sold
        uint256 leftover = sale.amountOfTokensToSell.sub(sale.totalTokensSold);

        if (leftover > 0) {
            sale.token.safeTransfer(msg.sender, leftover);
        }
    }

    /// @notice     Check signature user submits for registration.
    /// @notice     检查注册签名函数
    /// @param      signature is the message signed by the trusted entity (backend)
    /// @param      user is the address of user which is registering for sale
    function checkRegistrationSignature(
        bytes memory signature,
        address user
    ) public view returns (bool) {
        bytes32 hash = keccak256(
            abi.encodePacked(user, address(this))
        );
        bytes32 messageHash = hash.toEthSignedMessageHash();
        return admin.isAdmin(messageHash.recover(signature));
    }

    // Function to check if admin was the message signer
    // 检查参与签名函数
    function checkParticipationSignature(
        bytes memory signature,
        address user,
        uint256 amount
    ) public view returns (bool) {
        return
        admin.isAdmin(
            getParticipationSigner(
                signature,
                user,
                amount
            )
        );
    }

    /// @notice     Check who signed the message
    /// @notice     获取参与签名人函数
    /// @param      signature is the message allowing user to participate in sale
    /// @param      user is the address of user for which we're signing the message
    /// @param      amount is the maximal amount of tokens user can buy
    function getParticipationSigner(
        bytes memory signature,
        address user,
        uint256 amount
    ) public view returns (address) {
        bytes32 hash = keccak256(
            abi.encodePacked(
                user,
                amount,
                address(this)
            )
        );
        bytes32 messageHash = hash.toEthSignedMessageHash();
        return messageHash.recover(signature);
    }

    /// @notice     Function to get participation for passed user address
    /// @notice     获取购买者信息
    function getParticipation(address _user)
    external
    view
    returns (
        uint256,
        uint256,
        uint256,
        bool[] memory
    )
    {
        Participation memory p = userToParticipation[_user];
        return (
        p.amountBought,
        p.amountETHPaid,
        p.timeParticipated,
        p.isPortionWithdrawn
        );
    }

    /// @notice     Function to get number of registered users for sale
    /// @notice     获取注册购买者数量函数
    function getNumberOfRegisteredUsers() external view returns (uint256) {
        return registration.numberOfRegistrants;
    }

    /// @notice     Function to get all info about vesting.
    /// @notice     获取释放信息函数
    function getVestingInfo()
    external
    view
    returns (uint256[] memory, uint256[] memory)
    {
        return (vestingPortionsUnlockTime, vestingPercentPerPortion);
    }

    // Function to act as a fallback and handle receiving ETH.
    // 接收 ETH 的回退函数
    receive() external payable {

    }
}
