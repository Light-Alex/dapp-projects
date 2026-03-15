// logs.ts
import { JsonRpcProvider, id } from "ethers";
const RPC_URL = "https://sepolia.infura.io/v3/your_key"; // 或 Alchemy

// ERC20 Transfer(address,address,uint256) 事件的 topic0
const TRANSFER_TOPIC0 = id("Transfer(address,address,uint256)");
const USDC_SEPOLIA = "0x..."; // 替换为目标合约

async function main() {
  const provider = new JsonRpcProvider(RPC_URL);
  const fromBlock = "0x0"; // 或指定最近区块
  const toBlock = "latest";
  const logs = await provider.send("eth_getLogs", [{
    address: USDC_SEPOLIA,
    fromBlock,
    toBlock,
    topics: [TRANSFER_TOPIC0]
  }]);
  console.log("logs:", logs.length);
}
main();