GoVM [![Build Status](https://travis-ci.org/lnsp/govm.svg?branch=master)](https://travis-ci.org/lnsp/govm)
=========

## Specification

- 16-Bit address space (from `0x0000` to `0xFFFF`)
- 8 registers
	- Operation registers (AX, BX, CX, DX)
	- Pointer registers (CP, SP)
	- Flag registers (ZF, CF)
- Interrupt pointers (IX)
	- On-Off interrupt
	- Keyboard interrupt
	- Stack overflow interrupt
- Big endian memory layout
- All standard operations supported
- Virtual console display (80x24 character grid, 16 colors)

## Memory layout
### `0 - F`
|  Address  |    Description    | Name |
|-----------|-------------------|------|
| `0`       | code pointer      | CP   |
| `2`       | stack pointer     | SP   |
| `4`       | zero flag         | ZF   |
| `6`       | carry flag        | CF   |
| `8`       | register          | AX   |
| `A`       | register          | BX   |
| `C`       | register          | CX   |
| `E`       | register          | DX   |

### `10 - FF`
|  Address  |    Description    | Name |
|-----------|-------------------|------|
| `10`      | interrupt value   | IX   |
| `12`      | ir state          | IT   |
| `14`      | ir keyboard       | IK   |
| `16`      | ir stack overflow | IS   |

### `100 - FFF`
|  Address  |    Description    | Name |
|-----------|-------------------|------|
| `100`     | stack base        | SB   |
| ...       | stack             | -    |
| `FFF`     | stack max         | -    |

### `1000 - 1FFFF`
The memory layout of this region differs from mode to mode. The graphics mode is stored in `1FFE`.

The only available mode at this time is the **terminal mode** (80x24).
In this mode the region (`1000` - `1EFF`) stores character and color data.
The first byte maps the color (first half foreground, second half background).
The second byte maps the character. The region (`1F00` - `1FFD`) stores 16-bit colors.

## License

Copyright 2016 Lennart Espe. All rights reserved.

Use of this source code is governed by a MIT-style
license that can be found in the LICENSE.md file.
