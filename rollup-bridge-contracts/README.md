# Nil L1-Contracts

This smart-contracts module contains contracts required for communication with L1.
Solidity smart contracts in the repository are used for proof verification and state root updates of L2 on the L1 chain

1. NilRollup
2. NilVerifier

## NilRollup

- NilRollup contract is the entrypoint contract for batch commits and proof verifications initiated by syncCommittee

- NilRollup contract contains 2 main functions:
   1. [commitBatch](./contracts/NilRollup.sol#L343)
   2. [updateState](./contracts/NilRollup.sol#L392)

## NilVerifier

1. [NilVerifier-Contract](./contracts/verifier/NilVerifier.sol) contains the [verify](./contracts/verifier/NilVerifier.sol#L9) proof logic
2. NilVerifier contract is non-upgradeable and stateless contract

## Local Development - compilation

1. copy `.env.example` to `.env`
2. compile contracts:

```sh
npx hardhat clean && npx hardhat compile
```

### For nix pipeline run:

- set all pre-requisite variables in .env
```
GETH_RPC_ENDPOINT="http://localhost:8545"
GETH_PRIVATE_KEY=""
GETH_WALLET_ADDRESS=""
```

### Deploy contracts to geth instance via nix script:

- ensure to include this command as part of nix script to deploy the contracts to geth instance

```shell
npx hardhat deploy --network geth --tags NilContracts
```
