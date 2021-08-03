/* eslint-env mocha */

import fs from 'fs'
import path from 'path'
import chai from 'chai'
import { sha256 } from 'multiformats/hashes/sha2'
import * as Block from 'multiformats/block'
import { codecs } from './codecs.js'

const { assert } = chai
const fixturesDir = new URL('../fixtures/', import.meta.url)

describe('Codec fixtures', () => {
  for (const dir of fs.readdirSync(fixturesDir)) {
    const dirUrl = new URL(`./${dir}/`, fixturesDir)
    const stat = fs.statSync(dirUrl)
    if (!stat.isDirectory()) {
      continue
    }
    it(dir, async () => {
      const data = await loadFixture(dirUrl)
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

async function loadFixture (dir) {
  const data = {}
  for (const file of await fs.promises.readdir(dir)) {
    const ext = path.extname(file).slice(1)
    const cid = file.substring(0, file.length - ext.length - 1)
    const bytes = await fs.promises.readFile(new URL(file, dir))
    data[ext] = { cid, bytes }
  }
  return data
}
