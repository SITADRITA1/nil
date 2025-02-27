# Async calls and default tokens

On =nil;, contracts deployed on different shards can [**use async calls**](https://docs.nil.foundation/nil/smart-contracts/handling-async-execution) to communicate with each other.

This is typically done importing the *Nil.sol* contract and using the *Nil.asyncCall()* function:

```solidity
function asyncCall(
    address dst,
    address bounceTo,
    uint value,
    bytes memory callData
) {}
```

Since *Nil.asyncCall()* includes the *value* argument, this function can be used to [**pass default tokens**](https://docs.nil.foundation/nil/smart-contracts/tokens) between contracts on different shards.

## Task

This tutorial includes two contracts:

* *Caller*
* *Receiver*

The goal of *Caller* is to send some default tokens to *Receiver* by invoking the *sendValue()* function. *Caller* should also be able to receive tokens.

The goal of *Receiver* is to receive tokens sent from *Caller* and to send a small arbitrary portion of them back to *Caller*.

To complete this tutorial:

* Complete the *Caller* contract so that *sendValue()* sends funds to *Receiver* by calling the *depositAndReturn()* function. Also, make *Caller* be able to receive tokens
* Complete the *Receiver* contract so that it can receive tokens and send some of them back inside the *depositAndReturn()* function.

## Checks

This tutorial is verified once the following checks are passed:

1. *Caller* and *Receiver* are deployed and compiled
2. *sendValue()* is successfully called inside *Caller*
3. *Receiver* receives tokens from *Caller*
4. *depositAndReturn()* is successfully called inside *Receiver* after *Caller* uses *sendValue()*
5. *Receiver()* sends some tokens back to *Caller* when executing *depositAndReturn()*
6. *Caller* successfully receives these tokens

To run these checks:

1. Compile both contracts
2. Click on 'Run Checks'