import fs from 'fs'
import path from 'path'

export const fixturesDir = new URL('../fixtures/', import.meta.url)

export async function loadFixture (dir) {
  const data = {}
  for (const file of await fs.promises.readdir(dir)) {
    const ext = path.extname(file).slice(1)
    const cid = file.substring(0, file.length - ext.length - 1)
    const bytes = await fs.promises.readFile(new URL(file, dir))
    data[ext] = { cid, bytes }
  }
  return data
}

export function * fixtureDirectories () {
  for (const name of fs.readdirSync(fixturesDir)) {
    const stat = fs.statSync(new URL(`./${name}`, fixturesDir))
    if (!stat.isDirectory()) {
      continue
    }
    const url = new URL(`./${name}/`, fixturesDir)
    yield { name, url }
  }
}
