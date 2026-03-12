// test/debug-test.js
const { expect } = require("chai");

describe("Debug Contract", function () {
  it("should log debug information", async function () {
    const DebugContract = await ethers.getContractFactory("DebugContract");
    const debugContract = await DebugContract.deploy();
    
    // 调用会生成日志的函数
    const tx = await debugContract.setValue(42);
    await tx.wait();
    
    // 调用复杂函数
    const result = await debugContract.complexFunction(10, 20);
    expect(result).to.equal(60n);
  });
  
  it("should handle errors with good stack traces", async function () {
    const DebugContract = await ethers.getContractFactory("DebugContract");
    const debugContract = await DebugContract.deploy();
    
    // await expect(debugContract.complexFunction(2n**255n, 2n**255n))
    //   .to.be.reverted;
    await expect(debugContract.complexFunction(2n**255n, 2n**255n))
      .to.be.revertedWithPanic(0x11);  // Arithmetic overflow
  });
});