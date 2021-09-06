import fs from 'fs'
import { parse, patch, toString } from 'testmark.js'
import { codecs } from './codecs.js'
import { fixtureDirectories, loadFixture } from './util.js'

async function makeTestmark () {
  if (process.argv.length === 2) {
    throw new Error('Usage `make-testmark.js <path/to/ipld/metarepo>')
  }
  const repoRoot = new URL(process.argv[2], import.meta.url)

  const fixtures = []

  for (const { name, url } of fixtureDirectories()) {
    fixtures.push([name, await loadFixture(url)])
  }

  for (const codec of Object.keys(codecs)) {
    const dir = new URL(`specs/codecs/${codec}/fixtures/cross-codec/`, repoRoot)
    let stat
    try {
      stat = await fs.promises.stat(dir)
    } catch (err) {
      console.error(`Directory does not exist: ${dir.pathname}`)
      process.exit(1)
    }
    if (!stat.isDirectory()) {
      console.error(`Is not a directory: ${dir.pathname}`)
      process.exit(1)
    }
    const fixtureFile = new URL('index.md', dir)
    let tmDoc = parse(`# Cross-codec fixtures for ${codec}\n`)
    try {
      stat = await fs.promises.stat(fixtureFile)
      if (!stat.isFile()) {
        console.error(`Is not a file: ${fixtureFile.pathname}`)
        process.exit(1)
      }
      const contents = await fs.promises.readFile(fixtureFile, 'utf8')
      tmDoc = parse(contents)
    } catch (err) {
    }

    const patchHunks = []
    for (let [name, fixture] of fixtures) {
      if (fixture[codec]) {
        name = name.replace(/\s/g, '__')
        if (tmDoc.hunksByName.has(name)) { // patch it
          patchHunks.push({
            name,
            blockTag: '',
            body: fixture[codec].bytes.toString('hex')
          })
          patchHunks.push({
            name: `${name}-cid-${codec}`,
            blockTag: '',
            body: fixture[codec].cid
          })
          if (codec === 'dag-json') {
            // special case for dag-json since a human should be able to read it
            patchHunks.push({
              name: `${name}-string-form`,
              blockTag: 'json',
              body: new TextDecoder().decode(fixture[codec].bytes)
            })
          }
          for (const altCodec of Object.keys(fixture)) {
            if (altCodec === codec) {
              continue
            }
            patchHunks.push({
              name: `${name}-cid-${altCodec}`,
              blockTag: '',
              body: fixture[altCodec].cid
            })
          }
        } else { // create it
          let section = `## ${name}

### Bytes
[testmark]:# (${name})
\`\`\`
${fixture[codec].bytes.toString('hex')}
\`\`\`
`
          if (codec === 'dag-json') {
            // special case for dag-json since a human should be able to read it
            section += `
### String form
[testmark]:# (${name}-string-form)
\`\`\`
${new TextDecoder().decode(fixture[codec].bytes)}
\`\`\`
`
          }

          section += `
### ${codec} CID
[testmark]:# (${name}-cid-${codec})
\`\`\`
${fixture[codec].cid}
\`\`\`
`
          for (const altCodec of Object.keys(fixture)) {
            if (altCodec === codec) {
              continue
            }
            section += `
### ${altCodec} CID
[testmark]:# (${name}-cid-${altCodec})
\`\`\`
${fixture[altCodec].cid}
\`\`\`
`
          }
          tmDoc.lines = tmDoc.lines.concat(section.split('\n'))
        }
      }
    }
    tmDoc = patch(tmDoc, patchHunks)
    await fs.promises.writeFile(fixtureFile, toString(tmDoc), 'utf8')
  }
}

makeTestmark().catch((err) => {
  console.error(err)
  process.exit(1)
})
