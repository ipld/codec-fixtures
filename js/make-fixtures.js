import fs from 'fs/promises'
import path from 'path'

import { sha256 } from 'multiformats/hashes/sha2'
import * as Block from 'multiformats/block'
import { garbage } from 'ipld-garbage'
import { codecs } from './codecs.js'

const fixturesDir = new URL('../fixtures/', import.meta.url)
const fixturesSrcDir = new URL('../_fixtures_src/', import.meta.url)

async function makeGarbage () {
  const count = 25
  for (let i = 0; i < count;) {
    const value = garbage(5000)
    const block = await Block.encode({ value, codec: codecs['dag-cbor'].codec, hasher: sha256 })
    if (block.bytes.length < 1000) {
      continue
    }
    await fs.writeFile(new URL(`./garbage-${i.toString().padStart(2, '0')}.dag-cbor`, fixturesSrcDir), block.bytes)
    i++
  }
  return count
}

async function makeFixtures () {
  let count = 0
  await Promise.all((await fs.readdir(fixturesSrcDir)).map(async (file) => {
    const furl = new URL(file, fixturesSrcDir)
    const stat = await fs.stat(furl)
    if (!stat.isFile()) {
      return
    }
    const ext = path.extname(file).slice(1)
    if (!codecs[ext]) {
      console.error(`Unknown extension for file '${file}'`)
      return
    }
    const name = file.substring(0, file.length - ext.length - 1)
    const bytes = await fs.readFile(furl)
    let value
    try {
      value = codecs[ext].codec.decode(bytes)
    } catch (err) {
      console.error(`Failed to decode fixture ${file}`)
      throw err
    }
    const fdir = new URL(`./${name}/`, fixturesDir)
    try {
      await fs.mkdir(fdir)
    } catch (err) {
      if (err.code !== 'EEXIST') {
        throw err
      }
    }
    for (const { codec, complete } of Object.values(codecs)) {
      let block
      try {
        block = await Block.encode({ value, codec, hasher: sha256 })
      } catch (err) {
        if (!complete) { // failure is allowed, this codec can't handle this form
          continue
        }
        throw err
      }
      await fs.writeFile(new URL(`./${block.cid.toString()}.${codec.name}`, fdir), block.bytes)
      count++
    }
  }))
  return count
}

const p = process.argv.includes('--garbage') ? makeGarbage() : makeFixtures()
p.then((count) => {
  console.log(`Wrote ${count} fixtures`)
}).catch((err) => {
  console.error(err)
  process.exit(1)
})
