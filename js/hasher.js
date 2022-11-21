import { keccak256 as k } from 'ethereum-cryptography/keccak.js'
import { from } from 'multiformats/hashes/hasher'
import { coerce } from 'multiformats/bytes'

export const keccak256 = from({
    name: 'keccak-256',
    code: 0x1b,
    encode: (input) => coerce(k(input))
})
