//SPDX-License-Identifier: UNLICENSED
pragma solidity 0.6.12;

// 管理员合约
// 实现基于角色的访问控制（RBAC），管理管理员列表
contract Admin {

    // Listing all admins
    // 所有管理员地址数组
    address [] public admins;

    // Modifier for easier checking if user is admin
    // 管理员映射：用于快速检查用户是否为管理员
    mapping(address => bool) public isAdmin;

    // Modifier restricting access to only admin
    // 修饰符：仅管理员可调用
    modifier onlyAdmin {
        require(isAdmin[msg.sender], "Only admin can call.");
        _;
    }

    // Constructor to set initial admins during deployment
    // 构造函数：部署时设置初始管理员列表
    constructor (address [] memory _admins) public {
        for(uint i = 0; i < _admins.length; i++) {
            admins.push(_admins[i]);
            isAdmin[_admins[i]] = true;
        }
    }

    // 添加管理员函数
    function addAdmin(
        address _adminAddress
    )
    external
    onlyAdmin
    {
        // Can't add 0x address as an admin
        // 不能添加零地址作为管理员
        require(_adminAddress != address(0x0), "[RBAC] : Admin must be != than 0x0 address");
        // Can't add existing admin
        // 不能添加已存在的管理员
        require(!isAdmin[_adminAddress], "[RBAC] : Admin already exists.");
        // Add admin to array of admins
        // 将管理员添加到数组
        admins.push(_adminAddress);
        // Set mapping
        // 设置映射
        isAdmin[_adminAddress] = true;
    }

    // 移除管理员函数
    function removeAdmin(
        address _adminAddress
    )
    external
    onlyAdmin
    {
        // Admin has to exist
        // 管理员必须存在
        require(isAdmin[_adminAddress]);
        // 至少保留一个管理员，否则合约将无法使用
        require(admins.length > 1, "Can not remove all admins since contract becomes unusable.");
        uint i = 0;

        // 查找要移除的管理员在数组中的位置
        while(admins[i] != _adminAddress) {
            if(i == admins.length) {
                revert("Passed admin address does not exist");
            }
            i++;
        }

        // Copy the last admin position to the current index
        // 将最后一个管理员的位置复制到当前索引（替换要删除的）
        admins[i] = admins[admins.length-1];

        // 设置映射为 false
        isAdmin[_adminAddress] = false;

        // Remove the last admin, since it's double present
        // 移除最后一个管理员（因为它已经复制到被删除的位置）
        admins.pop();
    }

    // Fetch all admins
    // 获取所有管理员列表
    function getAllAdmins()
    external
    view
    returns (address [] memory)
    {
        return admins;
    }

}
