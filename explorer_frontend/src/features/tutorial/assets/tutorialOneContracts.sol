// SPDX-License-Identifier: MIT

pragma solidity ^0.8.9;

import "@nilfoundation/smart-contracts/contracts/Nil.sol";

// read more:
// https://docs.nil.foundation/nil/key-principles/async-execution
// https://docs.nil.foundation/nil/smart-contracts/handling-async-execution/

/**
 * @title Caller
 * @author =nil; Foundation
 * @notice The Caller contract must use sendValue() to send some
 * default tokens to the Receiver contract.
 * Caller must also be able to receive default tokens.
 */
contract Caller {
    using Nil for address;

    // Should send some default tokens to the Receiver contract
    // using Nil.asyncCall().
    function sendValue(address dst) public {}
}

/**
 * @title Receiver
 * @author =nil; Foundation
 * @notice The Receiver contract must be able to receive default tokens
 * when the depositAndReturn() function is called.

 */
contract Counter {
    function depositAndReturn() public {}
}
