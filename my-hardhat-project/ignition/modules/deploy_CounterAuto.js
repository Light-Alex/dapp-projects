const { buildModule } = require("@nomicfoundation/hardhat-ignition/modules");

module.exports = buildModule("IntervalCounter", (m) => {
  const intervalCounter = m.contract("IntervalCounter", ["10"]);
  return { intervalCounter };
});