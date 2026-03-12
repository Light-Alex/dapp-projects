// test/MyToken.test.js
const { expect } = require("chai");
const { ethers } = require("hardhat");

describe("MyToken Contract", function () {
  let MyToken;
  let myToken;
  let owner;
  let addr1;
  let addr2;
  let addrs;

  beforeEach(async function () {
    // 获取合约工厂
    MyToken = await ethers.getContractFactory("MyToken");
    
    // 获取签名者
    [owner, addr1, addr2, ...addrs] = await ethers.getSigners();

    // 部署合约
    myToken = await MyToken.deploy(
      "My Token",
      "MTK",
      1000000,  // 初始供应量
      18        // 小数位数
    );
    
    await myToken.waitForDeployment();
  });

  describe("Deployment", function () {
    it("Should set the right owner", async function () {
      expect(await myToken.owner()).to.equal(owner.address);
    });

    it("Should assign the total supply of tokens to the owner", async function () {
      const ownerBalance = await myToken.balanceOf(owner.address);
      expect(await myToken.totalSupply()).to.equal(ownerBalance);
    });

    it("Should have correct decimals", async function () {
      expect(await myToken.decimals()).to.equal(18);
    });
  });

  describe("Transactions", function () {
    it("Should transfer tokens between accounts", async function () {
      // 转账50个代币从owner到addr1
      await myToken.transfer(addr1.address, 50);
      const addr1Balance = await myToken.balanceOf(addr1.address);
      expect(addr1Balance).to.equal(50);

      // 从addr1转账50个代币到addr2
      await myToken.connect(addr1).transfer(addr2.address, 50);
      const addr2Balance = await myToken.balanceOf(addr2.address);
      expect(addr2Balance).to.equal(50);
    });

    it("Should fail if sender doesn't have enough tokens", async function () {
      const initialOwnerBalance = await myToken.balanceOf(owner.address);
      
      // 尝试从addr1转账1个代币到owner（应该失败）
      await expect(
        myToken.connect(addr1).transfer(owner.address, 1)
      ).to.be.reverted;

      // owner的余额不应该改变
      expect(await myToken.balanceOf(owner.address)).to.equal(
        initialOwnerBalance
      );
    });

    it("Should update balances after transfers", async function () {
      const initialOwnerBalance = await myToken.balanceOf(owner.address);

      // 转账100个代币从owner到addr1
      await myToken.transfer(addr1.address, 100);

      // 转账50个代币从owner到addr2
      await myToken.transfer(addr2.address, 50);

      // 检查余额
      const finalOwnerBalance = await myToken.balanceOf(owner.address);
      expect(finalOwnerBalance).to.equal(initialOwnerBalance - BigInt(150));

      const addr1Balance = await myToken.balanceOf(addr1.address);
      expect(addr1Balance).to.equal(BigInt(100));

      const addr2Balance = await myToken.balanceOf(addr2.address);
      expect(addr2Balance).to.equal(BigInt(50));
    });
  });

  describe("Minting", function () {
    it("Should mint new tokens", async function () {
      const initialSupply = await myToken.totalSupply();
      const mintAmount = BigInt(1000);
      
      await myToken.mint(addr1.address, mintAmount);
      
      expect(await myToken.totalSupply()).to.equal(initialSupply + mintAmount);
      expect(await myToken.balanceOf(addr1.address)).to.equal(mintAmount);
    });

    it("Should fail if non-owner tries to mint", async function () {
      await expect(
        myToken.connect(addr1).mint(addr1.address, 1000)
      ).to.be.reverted;
    });
  });
});
