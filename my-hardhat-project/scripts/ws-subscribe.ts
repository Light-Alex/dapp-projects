// ws-subscribe.ts
import { WebSocketProvider, id } from "ethers";

const WSS = "wss://eth-sepolia.g.alchemy.com/v2/your_key"; // 或 Infura WSS
const TRANSFER_TOPIC0 = id("Transfer(address,address,uint256)");
const USDC_SEPOLIA = "0x...";

async function main() {
  const ws = new WebSocketProvider(WSS);

  ws.on("block", (n) => console.log("new block:", n));

  ws.on({
    address: USDC_SEPOLIA,
    topics: [TRANSFER_TOPIC0]
  }, (log) => {
    console.log("log:", log.transactionHash, log.blockNumber);
  });
}
main();