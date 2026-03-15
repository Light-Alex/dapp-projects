import { JsonRpcProvider, formatEther } from "ethers";

const ALCHEMY_URL = "https://eth-mainnet.g.alchemy.com/v2/QyrReS0lko6EW06lqxBYJ2N5Xxwwi5UN";
const INFURA_URL  = "https://sepolia.infura.io/v3/your_key";

async function main() {
  const provider = new JsonRpcProvider(ALCHEMY_URL); // 或 INFURA_URL
  const [block, balance] = await Promise.all([
    provider.getBlockNumber(),
    provider.getBalance("vitalik.eth") // ENS 也可解析
  ]);
  console.log("block", block);
  console.log("balance(ETH)", formatEther(balance));
}
main();