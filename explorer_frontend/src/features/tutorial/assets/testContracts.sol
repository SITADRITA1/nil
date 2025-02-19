// SPDX-License-Identifier: MIT

pragma solidity ^0.8.9;

import "@nilfoundation/smart-contracts/contracts/Nil.sol";

// Caller contract is a simple proxy
// It demonstrates how to interact with another contract (Counter)
// Located (possibly) on a different shard
// In this case delivery of either response or error is guaranteed by the system

// To test, deploy Caller and Counter on separate shards
// Try executing `Caller.call` 5 times

// read more:
// https://docs.nil.foundation/nil/key-principles/async-execution
// https://docs.nil.foundation/nil/smart-contracts/handling-async-execution/#retreiving-values

contract Caller {
    using Nil for address;

    // Sends an async request to the Counter contract to invoke the increment method
    // It's guaranteed by the system that either response or error will be returned
    function call(address dst) public returns (uint256) {
        bytes memory val;
        bool ok;
        (val, ok) = Nil.awaitCall(
            dst, // Address of the destination contract (Counter)
            Nil.ASYNC_REQUEST_MIN_GAS, // Amount of gas reserved to process the response
            abi.encodeWithSignature("increment()") // Encoded signature of the increment function
        );

        // the request can fail on the destination contract (e.g. because OutOfGas error)
        require(ok == true, "Failed to perform an async request");

        // extract returned value if the request was successful
        return abi.decode(val, (uint256));
    }
}

// Counter contract is a simple stateful contract that keeps track of a counter value
// It provides method to increment the value

contract Counter {
    uint256 private value; // Stores the current counter value

    // Increments the counter by 1 and returns its value
    function increment() public returns (uint256) {
        value += 1;

        // dummy condition to illustrate error response after several calls
        require(value < 5, "Limit reached");

        return value;
    }
}
