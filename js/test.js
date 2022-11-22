/* eslint-env mocha */

import chai from 'chai'
import { sha256 } from 'multiformats/hashes/sha2'
import * as Block from 'multiformats/block'
import { codecs, ethCodecs } from './codecs.js'
import { keccak256 } from './hasher.js'
import {fixtureDirectories, keccak256FixtureDirectories, loadFixture} from './util.js'

const { assert } = chai

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

describe('Ethereum codec fixtures', () => {
  for (const { name, url } of keccak256FixtureDirectories()) {
    it (name, async () => {
      const data = await loadFixture(url)
      for (const [fromCodec, { bytes }] of Object.entries(data)) {
        const value = ethCodecs[fromCodec].codec.decode(bytes)
        for (const [toCodec, { cid }] of Object.entries(data)) {
          const block = await Block.encode({ value, codec: ethCodecs[toCodec].codec, hasher: keccak256})
          assert.equal(block.cid.toString(), cid, `CIDs match for data decoded from ${fromCodec} encoded as ${toCodec}`)
        }
      }
    })
  }
})
