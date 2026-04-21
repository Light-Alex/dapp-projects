const { ethers, upgrades } = require("hardhat");
const { Side, SaleKind } = require("../test/common");
const { toBn } = require("evm-bn");

/**  * 2024/12/22 in sepolia testnet
 * esVault contract deployed to: 0x75EC7448bC37c1FB484520C45b40F1564eBd0d19
     esVault ImplementationAddress: 
     esVault AdminAddress: 
   esDex contract deployed to: 0x5560e1c2E0260c2274e400d80C30CDC4B92dC8ac
      esDex ImplementationAddress: 
      esDex AdminAddress: 
 */

const esDex_name = "EasySwapOrderBook";
const esDex_address = "0xCf7Ed3AccA5a467e9e704C703E8D87F634fB0Fc9";
// const esDex_address = "0xcEE5AA84032D4a53a0F9d2c33F36701c3eAD5895"

const esVault_name = "EasySwapVault";
const esVault_address = "0xe7f1725E7734CE288F8367e1Bb143E90bb3F0512";
// const esVault_address = "0xaD65f3dEac0Fa9Af4eeDC96E95574AEaba6A2834"

const erc721_name = "TestERC721";
const erc721_address = "0x5FbDB2315678afecb367f032d93F642f64180aa3";
// const erc721_address = "0xF2e0BA02a187F19F5A390E4f990c684d81A833A0"

let esDex, esVault, testERC721;
let deployer, signer;
async function main() {
  [deployer, trader] = await ethers.getSigners();
  signer = await ethers.provider.getSigner();
  console.log("signer: ", signer.address);
  console.log("deployer: ", deployer.address);
  console.log("trader: ", trader.address);
  // esVault = await ethers.deployContract(esVault_name, []);
  // 部署 EasySwapVault 合约
  const EasySwapVault = await ethers.getContractFactory(esVault_name);
  esVault = await upgrades.deployProxy(EasySwapVault, {
    initializer: "initialize",
  });
  console.log("EasySwapVault deployed to:", await esVault.getAddress());

  const owner = await esVault.owner();
  console.log("Vault owner:", owner);

  EasySwapOrderBook = await ethers.getContractFactory(esDex_name);
  esDex = await upgrades.deployProxy(
    EasySwapOrderBook,
    [
      200, // newProtocolShare (默认协议费)
      await esVault.getAddress(), // newVault (EasySwapVault 地址)
      esDex_name, // EIP712Name
      "1", // EIP712Version
    ],
    { initializer: "initialize" }
  );
  console.log("EasySwapOrderBook deployed to:", await esDex.getAddress());

  const esDexAddress = await esDex.getAddress();
  tx = await esVault.setOrderBook(esDexAddress);
  await tx.wait();
  console.log("esVault setOrderBook tx:", tx.hash);

  testERC721 = await ethers.deployContract(erc721_name, []);
  console.log("ERC721 deployed to:", await testERC721.getAddress());
  // esDex = await (
  //   await ethers.getContractFactory(esDex_name)
  // ).attach(esDex_address);

  // esVault = await (
  //   await ethers.getContractFactory(esVault_name)
  // ).attach(esVault_address);

  // testERC721 = await (
  //   await ethers.getContractFactory(erc721_name)
  // ).attach(erc721_address);

  // 1. setApprovalForAll
  await approvalForVault();
  await testERC721.mint(deployer.address, 0); // 设置NFT tokenId=0的所有者
  // 2. make order
  await testMakeOrder();

  for (let i = 1; i < 20; i++) {
    await testERC721.mint(deployer.address, i);
    await testMakeOrder(i);
  }

  // 3. cancel order
  // let orderKeys = [];
  // await testCancelOrder(orderKeys);

  // let orderKeys1 = ["0xa48c77f5aa25cd7b0d207b491cf7a0ef5cc5cf15e3c1f9534b6791ef856f0dbe"]
  // let orderKeys2 = ["0x2f01e4ef5cbea217934b2bb27a73fac35032a75ffb030dea41fdb995c55f3069",
  //     "0x3450ada942fc2595d7d12bd6385cf3f1b03a614b9076bb23adaf808205e49d3b"]

  // await testCancelOrder(orderKeys1);
  // await testCancelOrder(orderKeys2);

  // 4. match order
  // await testMatchOrder();

  // let orderKeys = ["0x98e25dd9a45bbf79100ebe3b1b311b2b6702a28c9fca5ee317feb0049893faa5",
  //     "0x0c78b81d5da49fe7fd13832aac4aba9f79f31d25453b61ed09ec3ce941adca70",
  //     "0x201dc11898ad0213485b4b34b9702beedc8f3bbcc71b2e38512508adb59c8ea9"];

  // for (let i = 0; i < 2; i++) {
  //     let info = await getOrderInfo(orderKeys[i]);
  //     let sellOrder = info.order;
  //     // console.log("sellOrder: ", sellOrder);
  //     let buyOrder = {
  //         side: Side.Bid,
  //         saleKind: SaleKind.FixedPriceForItem,
  //         maker: trader.address,
  //         nft: sellOrder.nft,
  //         price: sellOrder.price,
  //         expiry: sellOrder.expiry,
  //         salt: sellOrder.salt,
  //     }

  //     let tx = await esDex.connect(trader).matchOrder(sellOrder, buyOrder, { value: toBn("0.002") });
  //     let txRec = await tx.wait();
  //     console.log("matchOrder tx: ", tx.hash);
  // }

  // 5. else
  // await withdrawProtocolFee();
  // await testBatchTransferERC721();
}

async function approvalForVault() {
  // check is approved
  let isApproved = await testERC721.isApprovedForAll(
    deployer.address,
    await esVault.getAddress()
  );
  console.log("isApprovedForAll: ", isApproved);
  if (isApproved) {
    console.log("Already approved");
    return;
  }

  let tx = await testERC721.setApprovalForAll(await esVault.getAddress(), true);
  await tx.wait();
  console.log("Approval tx:", tx.hash);
}

async function testMakeOrder(tokenId = 0) {
  let now = parseInt(new Date() / 1000) + 100000;
  let salt = 1;
  let nftAddress = await testERC721.getAddress();
  let price = "0.002"; // 确保是字符串
  let bnPrice = ethers.parseUnits(price, 18); // 转换为 BigNumber，假设 18 位小数
  console.log("Price after:", bnPrice);

  // let tokenId = 0;
  let order = {
    side: Side.List,
    saleKind: SaleKind.FixedPriceForItem,
    maker: deployer.address,
    nft: [tokenId, nftAddress, 1],
    price: bnPrice,
    expiry: now,
    salt: salt,
  };
  console.log("nftAddress: ", nftAddress);
  // console.log("Price before toBn:", order.price);
  // let bnPrice = toBn(order.price);
  console.log("order:", order);

  tx = await esDex.makeOrders([order]);
  txRec = await tx.wait();
  console.log(tx.hash);
}

async function testCancelOrder(orderKeys) {
  tx = await esDex.cancelOrders(orderKeys);
  txRec = await tx.wait();
  console.log(txRec);
}

async function testMatchOrder() {
  let now = 1734937947;
  let salt = 1;
  let tokenId = 0;
  let nftAddress = erc721_address;

  let sellOrder = {
    side: Side.List,
    saleKind: SaleKind.FixedPriceForItem,
    maker: deployer.address,
    nft: [tokenId, nftAddress, 1],
    price: toBn("0.002"),
    expiry: now,
    salt: salt,
  };

  // tx = await esDex.makeOrders([sellOrder]);
  // txRec = await tx.wait();
  // console.log("sellOrder tx: ", tx.hash);

  // ====
  let buyOrder = {
    side: Side.Bid,
    saleKind: SaleKind.FixedPriceForCollection,
    maker: trader.address,
    nft: [tokenId, nftAddress, 1],
    price: toBn("0.002"),
    expiry: now,
    salt: salt,
  };

  tx = await esDex
    .connect(trader)
    .matchOrder(sellOrder, buyOrder, { value: toBn("0.002") });
  txRec = await tx.wait();
  console.log("matchOrder tx: ", txRec.hash);
}

async function testBatchTransferERC721() {
  toAddr = "0x7752A564c941f7145AdF8B50AA2eC975cEf58689";
  nftAddr = "0x3c8ac104dcbf03ae12c9ac80aa830e1b39609e97";
  tokenId = 1159;
  asset = [nftAddr, tokenId];
  assets = [asset];
  tx = await esVault.callStatic.batchTransferERC721(toAddr, assets);
  console.log("tx: ", tx);
}

async function getOrderInfo(orderKey) {
  orderInfo = await esDex.orders(orderKey);
  // console.log("orderInfo: ", orderInfo);
  return orderInfo;
}

async function getfillsStat(orderKey) {
  fillStat = await esDex.filledAmount(orderKey);
  // console.log(fillStat);
  return fillStat;
}

async function withdrawProtocolFee() {
  await esDex.withdrawETH(deployer.address, toBn("0.00011"), {
    gasLimit: 100000,
  });
  console.log("WithdrawETH succeed.");
}

// We recommend this pattern to be able to use async/await everywhere
// and properly handle errors.
main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
