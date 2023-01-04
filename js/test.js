/* eslint-env mocha */

import chai from 'chai'
import { sha256 } from 'multiformats/hashes/sha2'
import * as Block from 'multiformats/block'
import { codecs } from './codecs.js'
import {
  fixtureDirectories,
  negativeFixtureCodecs,
  negativeFixturesEncode,
  negativeFixturesDecode,
  loadFixture
} from './util.js'

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

describe.only('Codec negative fixtures', () => {
  for (const codec of negativeFixtureCodecs()) {
    describe(codec, () => {
      const { encode } = codecs[codec].codec
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
    })
  }
})
