import fs from 'fs'
import path from 'path'

export const fixturesDir = new URL('../fixtures/', import.meta.url)
export const negativeFixturesDir = new URL('../negative-fixtures/', import.meta.url)

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

function * iterate (type, ...dirs) {
  let part = `.`
  if (dirs.length > 1) {
    part = `${part}/${dirs.slice(1).join('/')}/`
  }
  const base = new URL(part, dirs[0])
  try {
    fs.statSync(base)
  } catch (e) {
    if (e.code == 'ENOENT') { // ignore missing
      return
    }
  }
  for (const name of fs.readdirSync(base)) {
    let url = new URL(`./${name}`, base)
    const stat = fs.statSync(url)
    if ((type === 'dir' && !stat.isDirectory()) || (type === 'file' && stat.isDirectory())) {
      continue
    }
    if (type === 'dir') {
      url = new URL(`./${name}/`, base)
    }
    yield { name, url }
  }
}

export function * fixtureDirectories () {
  yield * iterate('dir', fixturesDir)
}

export function * negativeFixtureCodecs () {
  for (const { name } of iterate('dir', negativeFixturesDir)) {
    yield name
  }
}

export function * negativeFixtures (type, codec) {
  for (const { name, url } of iterate('file', negativeFixturesDir, codec, type)) {
    const fixtureText = fs.readFileSync(url, 'utf8')
    yield JSON.parse(fixtureText)
  }
}

export function * negativeFixturesEncode (codec) {
  yield * negativeFixtures('encode', codec)
}

export function * negativeFixturesDecode (codec) {
  yield * negativeFixtures('decode', codec)
}
