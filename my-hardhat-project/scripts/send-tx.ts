import { JsonRpcProvider, Wallet, parseEther } from "ethers";

const RPC_URL = "https://eth-sepolia.g.alchemy.com/v2/your_key"; // 或 Infura
const PRIVATE_KEY = process.env.PRIVATE_KEY!; // 切勿硬编码

async function main() {
  const provider = new JsonRpcProvider(RPC_URL);
  const wallet = new Wallet(PRIVATE_KEY, provider);

  const tx = await wallet.sendTransaction({
    to: "0x0000000000000000000000000000000000000001",
    value: parseEther("0.001")
  });
  console.log("tx sent:", tx.hash);
  const receipt = await tx.wait();
  console.log("mined in block:", receipt?.blockNumber);
}
main();
