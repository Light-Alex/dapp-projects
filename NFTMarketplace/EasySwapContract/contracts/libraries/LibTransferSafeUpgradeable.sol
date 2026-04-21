// SPDX-License-Identifier: AGPL-3.0-only
pragma solidity ^0.8.19;
// 导入OpenZeppelin可升级合约的接口
import {IERC721} from "@openzeppelin/contracts-upgradeable/token/ERC721/ERC721Upgradeable.sol";
import {IERC20} from "@openzeppelin/contracts-upgradeable/token/ERC20/ERC20Upgradeable.sol";

/// @notice Safe ETH and ERC20 transfer library that gracefully handles missing return values.
/// @author Edit from solmate (https://github.com/transmissions11/solmate/blob/main/src/utils/SafeTransferLib.sol)
/// @dev Use with caution! Some functions in this library knowingly create dirty bits at the destination of the free memory pointer.
/// @dev Note that none of the functions in this library check that a token has code at all! That responsibility is delegated to the caller.
// 安全转账库（支持 ETH/ERC20/ERC721）
library LibTransferSafeUpgradeable {
    /*//////////////////////////////////////////////////////////////
                             ETH OPERATIONS
    //////////////////////////////////////////////////////////////*/
    // ETH 转账操作
    function safeTransferETH(address to, uint256 amount) internal {
        bool success;

        assembly {
            // 内联汇编实现ETH转账
            // Transfer the ETH and store if it succeeded or not.
            // 调用call函数转账ETH（参数顺序：gasLimit, to, value, 输入内存起始位置，输入长度，输出内存起始位置，输出长度）
            success := call(gas(), to, amount, 0, 0, 0, 0)
        }

        require(success, "ETH_TRANSFER_FAILED"); // 校验调用结果
    }

    /*//////////////////////////////////////////////////////////////
                            ERC20 OPERATIONS
    //////////////////////////////////////////////////////////////*/
    // ERC20 transferFrom 函数
    function safeTransferFrom(
        IERC20 token,
        address from,
        address to,
        uint256 amount
    ) internal {
        bool success;

        assembly {
            // Get a pointer to some free memory.
            let freeMemoryPointer := mload(0x40) // 获取空闲内存指针

            // Write the abi-encoded calldata into memory, beginning with the function selector.
            // 构造transferFrom函数签名和参数
            mstore(
                freeMemoryPointer,
                0x23b872dd00000000000000000000000000000000000000000000000000000000
            )
            mstore(add(freeMemoryPointer, 4), from) // 写入from地址（偏移4字节）Append the "from" argument.
            mstore(add(freeMemoryPointer, 36), to) // 写入to地址（偏移36字节）Append the "to" argument.
            mstore(add(freeMemoryPointer, 68), amount) // 写入转账金额（偏移68字节）Append the "amount" argument.
            // 执行外部调用并校验结果
            success := and(
                // Set success to whether the call reverted, if not we check it either
                // returned exactly 1 (can't just be non-zero data), or had no return data.
                or(
                    and(eq(mload(0), 1), gt(returndatasize(), 31)), // 检查返回数据是否为true
                    iszero(returndatasize()) // 允许没有返回数据（兼容老版本ERC20）
                ),
                // We use 100 because the length of our calldata totals up like so: 4 + 32 * 3.
                // We use 0 and 32 to copy up to 32 bytes of return data into the scratch space.
                // Counterintuitively, this call must be positioned second to the or() call in the
                // surrounding and() call or else returndatasize() will be zero during the computation.
                call(gas(), token, 0, freeMemoryPointer, 100, 0, 32) // 调用长度100字节（4+32*3）
            )
        }

        require(success, "TRANSFER_FROM_FAILED");
    }

    // ERC20 transfer 函数
    function safeTransfer(IERC20 token, address to, uint256 amount) internal {
        bool success;

        assembly {
            // Get a pointer to some free memory.
            let freeMemoryPointer := mload(0x40)
            // 构造transfer函数签名和参数
            // Write the abi-encoded calldata into memory, beginning with the function selector.
            mstore(
                freeMemoryPointer,
                0xa9059cbb00000000000000000000000000000000000000000000000000000000
            )
            mstore(add(freeMemoryPointer, 4), to) // 写入目标地址Append the "to" argument.
            mstore(add(freeMemoryPointer, 36), amount) // 写入转账金额Append the "amount" argument.

            success := and(
                // Set success to whether the call reverted, if not we check it either
                // returned exactly 1 (can't just be non-zero data), or had no return data.
                or(
                    and(eq(mload(0), 1), gt(returndatasize(), 31)),
                    iszero(returndatasize())
                ),
                // We use 68 because the length of our calldata totals up like so: 4 + 32 * 2.
                // We use 0 and 32 to copy up to 32 bytes of return data into the scratch space.
                // Counterintuitively, this call must be positioned second to the or() call in the
                // surrounding and() call or else returndatasize() will be zero during the computation.
                call(gas(), token, 0, freeMemoryPointer, 68, 0, 32) // 调用长度68字节（4+32*2）
            )
        }

        require(success, "TRANSFER_FAILED");
    }

    // ERC20 approve
    function safeApprove(IERC20 token, address to, uint256 amount) internal {
        bool success;

        assembly {
            // Get a pointer to some free memory.
            let freeMemoryPointer := mload(0x40)
            // 构造approve函数签名和参数
            // Write the abi-encoded calldata into memory, beginning with the function selector.
            mstore(
                freeMemoryPointer,
                0x095ea7b300000000000000000000000000000000000000000000000000000000
            )
            mstore(add(freeMemoryPointer, 4), to) // 写入被授权地址Append the "to" argument.
            mstore(add(freeMemoryPointer, 36), amount) // 写入授权金额Append the "amount" argument.

            success := and(
                // Set success to whether the call reverted, if not we check it either
                // returned exactly 1 (can't just be non-zero data), or had no return data.
                or(
                    and(eq(mload(0), 1), gt(returndatasize(), 31)),
                    iszero(returndatasize())
                ),
                // We use 68 because the length of our calldata totals up like so: 4 + 32 * 2.
                // We use 0 and 32 to copy up to 32 bytes of return data into the scratch space.
                // Counterintuitively, this call must be positioned second to the or() call in the
                // surrounding and() call or else returndatasize() will be zero during the computation.
                call(gas(), token, 0, freeMemoryPointer, 68, 0, 32)
            )
        }

        require(success, "APPROVE_FAILED");
    }

    // ERC721 safeTransferFrom
    function safeTransferNFT(
        IERC721 nft,
        address from,
        address to,
        uint256 tokenId
    ) internal {
        bool success;

        assembly {
            let freeMemoryPointer := mload(0x40)
            // 构造ERC721转账参数
            mstore(
                freeMemoryPointer,
                0x42842e0e00000000000000000000000000000000000000000000000000000000
            )
            mstore(add(freeMemoryPointer, 4), from) // 转出地址
            mstore(add(freeMemoryPointer, 36), to) // 接收地址
            mstore(add(freeMemoryPointer, 68), tokenId) // Token ID

            success := and(
                // Set success to whether the call reverted, if not we check it either
                // returned exactly 1 (can't just be non-zero data), or had no return data.
                or(
                    and(eq(mload(0), 1), gt(returndatasize(), 31)),
                    iszero(returndatasize())
                ),
                // We use 100 because the length of our calldata totals up like so: 4 + 32 * 3.
                // We use 0 and 32 to copy up to 32 bytes of return data into the scratch space.
                // Counterintuitively, this call must be positioned second to the or() call in the
                // surrounding and() call or else returndatasize() will be zero during the computation.
                call(gas(), nft, 0, freeMemoryPointer, 100, 0, 32) // 调用长度100字节
            )
        }

        require(success, "NFT_TRANSFER_FROM_FAILED");
    }

    // 批量转账ERC721
    function safeTransferNFTs(
        IERC721 nft,
        address from,
        address to,
        uint256[] memory tokenIds
    ) internal {
        bool success;
        uint256 amountNFT = tokenIds.length;
        // 循环处理每个Token ID
        for (uint256 i; i < amountNFT; ) {
            uint256 tokenId = tokenIds[i];

            assembly {
                let freeMemoryPointer := mload(0x40)
                // 与单个转账相同的参数构造
                mstore(
                    freeMemoryPointer,
                    0x42842e0e00000000000000000000000000000000000000000000000000000000
                )
                mstore(add(freeMemoryPointer, 4), from)
                mstore(add(freeMemoryPointer, 36), to)
                mstore(add(freeMemoryPointer, 68), tokenId)

                success := and(
                    // Set success to whether the call reverted, if not we check it either
                    // returned exactly 1 (can't just be non-zero data), or had no return data.
                    or(
                        and(eq(mload(0), 1), gt(returndatasize(), 31)),
                        iszero(returndatasize())
                    ),
                    // We use 100 because the length of our calldata totals up like so: 4 + 32 * 3.
                    // We use 0 and 32 to copy up to 32 bytes of return data into the scratch space.
                    // Counterintuitively, this call must be positioned second to the or() call in the
                    // surrounding and() call or else returndatasize() will be zero during the computation.
                    call(gas(), nft, 0, freeMemoryPointer, 100, 0, 32)
                )
            }

            require(success, "NFT_TRANSFER_FROM_FAILED");
            // 安全递增计数器（无溢出检查）
            unchecked {
                ++i;
            }
        }
    }
}
