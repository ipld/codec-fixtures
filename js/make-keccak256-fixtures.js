import fs from 'fs/promises'
import path from 'path'

import { keccak256 } from './hasher.js'
import * as Block from 'multiformats/block'
import { ethCodecs } from './codecs.js'

const fixturesDir = new URL('../keccak256_fixtures/', import.meta.url)
const fixturesSrcDir = new URL('../_keccak256_fixtures_src/', import.meta.url)

async function makeFixtures () {
    await Promise.all((await fs.readdir(fixturesSrcDir)).map(async (file) => {
        const furl = new URL(file, fixturesSrcDir)
        const stat = await fs.stat(furl)
        if (!stat.isFile()) {
            return
        }
        const ext = path.extname(file).slice(1)
        if (!ethCodecs[ext]) {
            console.error(`Unknown extension for file '${file}'`)
            return
        }
        const name = file.substring(0, file.length - ext.length - 1)
        const bytes = await fs.readFile(furl)
        let value
        let codec
        try {
            codec = ethCodecs[ext].codec
            value = ethCodecs[ext].codec.decode(bytes)
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
        let block
        try {
            block = await Block.encode({ value, codec, hasher: keccak256 })
        } catch (err) {
            throw err
        }
        await fs.writeFile(new URL(`./${block.cid.toString()}.${codec.name}`, fdir), block.bytes)
    }))
}

makeFixtures().catch((err) => {
    console.error(err)
    process.exit(1)
})
