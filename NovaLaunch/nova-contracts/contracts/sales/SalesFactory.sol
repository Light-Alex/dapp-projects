// "SPDX-License-Identifier: UNLICENSED"
pragma solidity 0.6.12;

// 引入接口和合约
import "../interfaces/IAdmin.sol";
import "./C2NSale.sol";


// 销售工厂合约
// 负责部署和管理 C2NSale 销售合约
contract SalesFactory {

    // 管理员合约接口
    IAdmin public admin;
    // 分配质押合约地址
    address public allocationStaking;

    // 记录销售合约是否由工厂创建（用于验证）
    mapping (address => bool) public isSaleCreatedThroughFactory;

    // 销售所有者地址到销售合约地址的映射
    mapping(address => address) public saleOwnerToSale;
    // 代币地址到销售合约地址的映射
    mapping(address => address) public tokenToSale;

    // Expose so query can be possible only by position as well
    // 所有已部署的销售合约地址数组
    address [] public allSales;

    // 事件定义
    event SaleDeployed(address saleContract);                                           // 销售合约部署事件
    event SaleOwnerAndTokenSetInFactory(address sale, address saleOwner, address saleToken); // 销售者和代币设置事件

    // 修饰符：仅管理员可调用
    modifier onlyAdmin {
        require(admin.isAdmin(msg.sender), "Only Admin can deploy sales");
        _;
    }

    // 构造函数：初始化管理员合约和质押合约地址
    constructor (address _adminContract, address _allocationStaking) public {
        admin = IAdmin(_adminContract);
        allocationStaking = _allocationStaking;
    }

    // Set allocation staking contract address.
    // 设置分配质押合约地址
    function setAllocationStaking(address _allocationStaking) public onlyAdmin {
        require(_allocationStaking != address(0));
        allocationStaking = _allocationStaking;
    }


    // 部署新的销售合约
    function deploySale()
    external
    onlyAdmin
    {
        // 创建新的 C2NSale 合约实例
        C2NSale sale = new C2NSale(address(admin), allocationStaking);

        // 标记该合约由工厂创建
        isSaleCreatedThroughFactory[address(sale)] = true;
        // 添加到所有销售列表
        allSales.push(address(sale));

        emit SaleDeployed(address(sale));
    }

    // Function to return number of pools deployed
    // 返回已部署的销售合约数量
    function getNumberOfSalesDeployed() external view returns (uint) {
        return allSales.length;
    }

    // Function
    // 获取最新部署的销售合约地址
    function getLastDeployedSale() external view returns (address) {
        //
        if(allSales.length > 0) {
            return allSales[allSales.length - 1];
        }
        return address(0);
    }


    // Function to get all sales
    // 获取指定索引范围内的销售合约地址列表
    function getAllSales(uint startIndex, uint endIndex) external view returns (address[] memory) {
        require(endIndex > startIndex, "Bad input");

        address[] memory sales = new address[](endIndex - startIndex);
        uint index = 0;

        for(uint i = startIndex; i < endIndex; i++) {
            sales[index] = allSales[i];
            index++;
        }

        return sales;
    }

}
