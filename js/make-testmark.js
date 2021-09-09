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
    let tmDoc = parse(`---
templateEngineOverride: md
---

# Cross-codec fixtures for ${codec}

## Introduction

_TODO_

## Fixtures
`)
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
        name = name.replace(/[\s/]/g, '__')
        if (tmDoc.hunksByName.has(`${name}/${codec}/bytes`)) { // patch it
          patchHunks.push({
            name: `${name}/${codec}/bytes`,
            blockTag: '',
            body: fixture[codec].bytes.toString('hex').replace(/(.{120})/g, '$1\n')
          })
          patchHunks.push({
            name: `${name}/${codec}/cid`,
            blockTag: '',
            body: fixture[codec].cid
          })
          if (codec === 'dag-json') {
            // special case for dag-json since a human should be able to read it
            patchHunks.push({
              name: `${name}/${codec}/string`,
              blockTag: 'json',
              body: new TextDecoder().decode(fixture[codec].bytes)
            })
          }
          for (const altCodec of Object.keys(fixture)) {
            if (altCodec === codec) {
              continue
            }
            patchHunks.push({
              name: `${name}/${altCodec}/cid`,
              blockTag: '',
              body: fixture[altCodec].cid
            })
          }
        } else { // create it
          let section = `### ${name}

**Bytes**

[testmark]:# (${name}/${codec}/bytes)
\`\`\`
${fixture[codec].bytes.toString('hex').replace(/(.{120})/g, '$1\n')}
\`\`\`
`
          if (codec === 'dag-json') {
            // special case for dag-json since a human should be able to read it
            section += `
**String form**

[testmark]:# (${name}/${codec}/string)
\`\`\`json
${new TextDecoder().decode(fixture[codec].bytes)}
\`\`\`
`
          }

          section += `
**${codec} CID**

[testmark]:# (${name}/${codec}/cid)
\`\`\`
${fixture[codec].cid}
\`\`\`
`
          for (const altCodec of Object.keys(fixture)) {
            if (altCodec === codec) {
              continue
            }
            section += `
**${altCodec} CID**

[testmark]:# (${name}/${altCodec}/cid)
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
