# 项目：NFTMarketplace

## Communication

- **语言**: 请使用中文进行所有交流、注释和文档。
- **术语**: 保持技术术语英文

## 约定
- 画图工具使用 `mermaid`

## 部署
### 智能合约
```bash
cd EasySwapContract
nvm use v20.20.0
npx hardhat node
npx hardhat run scripts/deploy.js
npx hardhat run scripts/deploy_721.js
npx hardhat run scripts/interact.js
```