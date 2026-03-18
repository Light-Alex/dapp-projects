// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.13;

import "forge-std/Script.sol";

abstract contract BaseScript is Script {
    address internal deployer;
    string internal mnemonic;
    uint256 internal deployerPrivateKey;

    // 设置部署配置
    function setUp() public virtual {
      // // 从环境变量MNEMONIC获取到部署者钱包的助记词
      // mnemonic = vm.envString("MNEMONIC");

      // // 从部署者钱包的助记词中获取到部署者地址
      // (deployer, ) = deriveRememberKey(mnemonic, 0);

      // 从环境变量PRIVATE_KEY获取到部署者钱包的私钥
      if (block.chainid == 31337) {
        deployerPrivateKey = vm.envUint("PRIVATE_KEY");
      } else{
        deployerPrivateKey = vm.envUint("SEPOLIA_PRIVATE_KEY");
      }

      // 从私钥中获取到部署者地址
      deployer = vm.addr(deployerPrivateKey);
    }

    // 保存合约部署信息到 JSON 文件
    function saveContract(string memory network, string memory name, address addr) public {
      string memory chainId = vm.toString(block.chainid);
      string memory json1 = "key";

      // 将合约地址序列化为 JSON 格式
      string memory finalJson =  vm.serializeAddress(json1, "address", addr);

      // 创建按网络分类的输出目录（如 deployments/mainnet/1/）
      string memory dirPath = string.concat("deployments/", network, "/", chainId, "/");

      // 如果目录不存在，则创建目录
      if (!vm.isDir(dirPath)) {
        vm.createDir(dirPath, true);
      }

      // 将合约信息写入到指定文件（如 MyContract.json）
      string memory filePath = string.concat(dirPath, name, ".json");

      // 将合约信息写入到指定文件
      vm.writeJson(finalJson, filePath); 
    }

    function getNetworkName(uint256 chainId) public pure returns (string memory) {
        if (chainId == 1) return "mainnet";
        if (chainId == 11155111) return "sepolia";
        if (chainId == 97) return "bsc_test";
        if (chainId == 56) return "bsc";
        if (chainId == 31337) return "localhost";
        return "unknown";
    }

    // 广播部署交易
    modifier broadcaster() {
        // 开始广播部署交易，使用部署者地址作为交易发送者
        // vm.startBroadcast(deployer);
        vm.startBroadcast(deployerPrivateKey);  // 直接用私钥部署不需要解锁钱包
        _;
        // 停止广播，结束部署交易
        vm.stopBroadcast();
    }
}
