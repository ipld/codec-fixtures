import { createWriteStream } from 'fs'
import { join } from 'path'
import { pipeline } from 'stream/promises'
import { CID } from 'multiformats'
import { CarWriter } from '@ipld/car'
import { fixtureDirectories, loadFixture } from './util.js'

const outFile = join(process.cwd(), '..', 'fixtures.car')
const outStream = createWriteStream(outFile)
const { writer, out } = await CarWriter.create([])
const pipe = pipeline(out, outStream)

for (const { name, url } of fixtureDirectories()) {
  const data = await loadFixture(url)
  for (const { cid, bytes } of Object.values(data)) {
    await writer.put({ cid: CID.parse(cid), bytes })
  }
}

await writer.close()
await pipe
console.log(`Wrote fixtures to ${outFile}`)