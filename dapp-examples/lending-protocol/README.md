## 🏦 Lending and Borrowing Protocol on =nil;

## 🔍 Overview

This repository contains an example decentralized application (dApp) showcasing a lending and borrowing protocol built on the =nil; blockchain. This example demonstrates how to leverage sharded smart contracts, asynchronous communication, and cross-shard interactions using various methods of `nil.sol`.

### ✨ Features

- 💰 **Deposit USDT and ETH** into a lending pool
- 🔐 **Borrow assets** based on collateral
- 💳 **Repay borrowed assets** seamlessly
- 📊 **Oracle-based price updates** for accurate valuations

### 🚀 Key Highlights

- 🧩 **Sharded Smart Contracts**: Efficient workload distribution across shards
- ⚡ **Asynchronous Communication**: Transaction execution with minimal bottlenecks
- 🔗 **Cross-Shard Interactions**: Smart contract coordination across different shards

---

## ⚙️ Prerequisites

Before working with this repository, ensure you have the following installed:

- 📌 [Node.js](https://nodejs.org/) (version 16 or higher recommended)
- 📦 [npm](https://www.npmjs.com/) (included with Node.js)
- 🔨 Hardhat for smart contract development
- 🌍 A =nil; testnet RPC endpoint
- 🔑 `.env` file with RPC and private key configuration

Check your installed versions with:

```sh
node -v
npm -v
```

---

## 📦 Installation

1. 📥 Clone the repository:
   ```sh
   git clone https://github.com/NilFoundation/nil.git
   ```
2. 📂 Navigate to the project root and install dependencies:
   ```sh
   cd lending-protocol
   npm install
   ```
3. 🏗️ Compile the smart contracts:
   ```sh
   npx hardhat compile
   ```
4. 🚀 Run the end-to-end lending workflow:
   ```sh
   npx hardhat run-lending-protocol
   ```
   This script deploys contracts across different shards, sets up accounts, deposits assets, borrows against collateral, and processes repayments.

---

## 📜 Understanding the `run-lending-protocol` Flow

This command executes the following steps:

1. 🏗 **Deploys contracts** across multiple shards
2. 👥 **Creates smart contract-based accounts**
3. 📊 **Sets and verifies oracle prices** for assets
4. 💸 **Funds accounts with USDT and ETH**
5. 🏦 **Deposits funds** into the lending pool
6. 🔄 **Initiates borrowing** of ETH against USDT
7. ✅ **Processes loan repayment**

---

## 🤝 Contribution

This project serves as an example, but contributions are welcome to improve and expand its functionality!

### 💡 How You Can Contribute:

- ✍️ **Enhance lending mechanisms** and introduce new features
- 🔍 **Optimize sharding efficiency** for better scalability
- 🛠 **Improve cross-shard execution and smart contract interactions**

📌 Check out our list of open issues: [Issue](https://github.com/NilFoundation/nil/issues).
📖 For detailed contribution guidelines, refer to [Contribution Guide](https://github.com/NilFoundation/nil/blob/main/CONTRIBUTION-GUIDE.md)

🚀 **Thank you for your support, and happy building!** 🎉
