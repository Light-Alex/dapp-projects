const { buildModule } = require("@nomicfoundation/hardhat-ignition/modules");

module.exports = buildModule("EthPriceFeedModule", (m) => {
  const ethPriceFeed = m.contract("EthPriceFeed", ["0x694AA1769357215DE4FAC081bf1f309aDC325306"]);
  return { ethPriceFeed };
});