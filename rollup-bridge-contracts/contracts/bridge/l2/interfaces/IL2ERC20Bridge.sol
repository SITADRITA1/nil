// SPDX-License-Identifier: MIT
pragma solidity 0.8.28;

interface IL2ERC20Bridge {
  /**********
   * Events *
   **********/

  /// @notice Emitted when ERC20 token is deposited from L1 to L2 and transfer to recipient.
  /// @param l1Token The address of the token in L1.
  /// @param l2Token The address of the token in L2.
  /// @param from The address of sender in L1.
  /// @param to The address of recipient in L2.
  /// @param amount The amount of token withdrawn from L1 to L2.
  /// @param data The optional calldata passed to recipient in L2.
  event FinalizeDepositERC20(
    address indexed l1Token,
    address indexed l2Token,
    address indexed from,
    address to,
    uint256 amount,
    bytes data
  );

  /*************************
   * Public View Functions *
   *************************/

  /// @notice Return the corresponding l1 token address given l2 token address.
  /// @param l2Token The address of l2 token.
  function getL1ERC20Address(address l2Token) external view returns (address);

  /// @notice Return the corresponding l2 token address given l1 token address.
  /// @param l1Token The address of l1 token.
  function getL2ERC20Address(address l1Token) external view returns (address);

  /*****************************
   * Public Mutating Functions *
   *****************************/

  /// @notice Complete a deposit from L1 to L2 and send fund to recipient's account in L2.
  /// @dev Make this function payable to handle WETH deposit/withdraw.
  ///      The function should only be called by L2ScrollMessenger.
  ///      The function should also only be called by L1ERC20Gateway in L1.
  /// @param l1Token The address of corresponding L1 token.
  /// @param l2Token The address of corresponding L2 token.
  /// @param from The address of account who deposits the token in L1.
  /// @param to The address of recipient in L2 to receive the token.
  /// @param amount The amount of the token to deposit.
  /// @param data Optional data to forward to recipient's account.
  function finalizeDepositERC20(
    address l1Token,
    address l2Token,
    address from,
    address to,
    uint256 amount,
    bytes calldata data
  ) external payable;
}
