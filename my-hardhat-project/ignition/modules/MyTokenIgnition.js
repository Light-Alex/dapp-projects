const { buildModule } = require("@nomicfoundation/hardhat-ignition/modules");

module.exports = buildModule("MyTokenModule", (m) => {
  const token = m.contract("MyToken", [
    "Hello Token",  // name
    "HT",          // symbol
    1000,          // initialSupply
    18             // decimals
  ]);

  return { token };
});