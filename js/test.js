/* eslint-env mocha */

import * as chai from 'chai'
import { sha256 } from 'multiformats/hashes/sha2'
import * as Block from 'multiformats/block'
import { codecs } from './codecs.js'
import {
  fixtureDirectories,
  negativeFixtureCodecs,
  negativeFixturesEncode,
  negativeFixturesDecode,
  noncanonicalFixtureCodecs,
  noncanonicalFixturesDecode,
  loadFixture
} from './util.js'
import { bytes } from 'multiformats'

const { assert } = chai
const utfEncoder = new TextEncoder()

describe('Codec fixtures', () => {
  for (const { name, url } of fixtureDirectories()) {
    it(name, async () => {
      const data = await loadFixture(url)
      for (const [fromCodec, { bytes }] of Object.entries(data)) {
        const value = codecs[fromCodec].codec.decode(bytes)
        for (const [toCodec, { cid }] of Object.entries(data)) {
          const block = await Block.encode({ value, codec: codecs[toCodec].codec, hasher: sha256 })
          assert.equal(block.cid.toString(), cid, `CIDs match for data decoded from ${fromCodec} encoded as ${toCodec}`)
        }
      }
    })
  }
})

describe('Codec negative fixtures', () => {
  for (const codec of negativeFixtureCodecs()) {
    describe(codec, () => {
      const { encode, decode } = codecs[codec].codec

      for (const fixtures of negativeFixturesEncode(codec)) {
        for (const fixture of fixtures) {
          it(fixture.name, () => {
            const { name, error } = fixture
            if (!'dag-json' in fixture) {
              // TODO: when we need it, probably hex decode for others
              assert.fail('can\'t deal with fixture that doesn\'t have dag-json input')
            }
            const obj = codecs['dag-json'].codec.decode(utfEncoder.encode(JSON.stringify(fixture['dag-json'])))
            try {
              encode(obj)
              assert.fail('did not error')
            } catch (e) {
              assert.strictEqual(e.message, error)
            }
          })
        }
      }

      for (const fixtures of negativeFixturesDecode(codec)) {
        for (const { name, hex, error } of fixtures) {
          it(name, () => {
            const byts = bytes.fromHex(hex)
            try {
              decode(byts)
              assert.fail('did not error')
            } catch (e) {
              assert.include(e.message, error)
            }
          })
        }
      }
    })
  }
})

// Non-canonical fixtures: blocks that decode successfully but are not in
// canonical form, so they can never round-trip byte-for-byte. The contract:
// decoding must succeed, the decoded value must equal the value decoded from
// the canonical bytes, and canonical re-encoding must produce the canonical
// CID.
//
// TODO: once codec implementations expose an opt-in encoder for alternate
// forms (e.g. the dag-pb Data-first field order proposed by IPIP-550,
// https://github.com/ipfs/specs/pull/550), promote these to full round-trip
// fixtures with a per-fixture encode hint.
describe('Codec noncanonical fixtures', () => {
  for (const codec of noncanonicalFixtureCodecs()) {
    describe(codec, () => {
      const { decode } = codecs[codec].codec
      for (const fixtures of noncanonicalFixturesDecode(codec)) {
        for (const { name, hex, canonicalHex, canonicalCid } of fixtures) {
          it(name, async () => {
            const value = decode(bytes.fromHex(hex))
            const canonicalValue = decode(bytes.fromHex(canonicalHex))
            assert.deepEqual(value, canonicalValue, 'noncanonical and canonical bytes decode to the same value')
            const block = await Block.encode({ value, codec: codecs[codec].codec, hasher: sha256 })
            assert.equal(block.cid.toString(), canonicalCid, 'canonical re-encode produces the canonical CID')
          })
        }
      }
    })
  }
})
