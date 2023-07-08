import * as dagCBOR from '@ipld/dag-cbor'
import * as dagJSON from '@ipld/dag-json'
import * as dagPB from '@ipld/dag-pb'
import * as dagEthHeader from '@vulcanize/dag-eth/dist/header/src/index.js'
import * as dagEthLog from '@vulcanize/dag-eth/dist/log/src/index.js'
import * as dagEthLogTrie from '@vulcanize/dag-eth/dist/log_trie/src/index.js'
import * as dagEthRct from '@vulcanize/dag-eth/dist/rct/src/index.js'
import * as dagEthRctTrie from '@vulcanize/dag-eth/dist/rct_trie/src/index.js'
import * as dagEthStateAccount from '@vulcanize/dag-eth/dist/state_account/src/index.js'
import * as dagEthStateTrie from '@vulcanize/dag-eth/dist/state_trie/src/index.js'
import * as dagEthStorageTrie from '@vulcanize/dag-eth/dist/storage_trie/src/index.js'
import * as dagEthTx from '@vulcanize/dag-eth/dist/tx/src/index.js'
import * as dagEthTxTrie from '@vulcanize/dag-eth/dist/tx_trie/src/index.js'
import * as dagEthUncles from '@vulcanize/dag-eth/dist/uncles/src/index.js'

export const codecs = {
  [dagCBOR.name]: { codec: dagCBOR, complete: true },
  [dagJSON.name]: { codec: dagJSON, complete: true },
  [dagPB.name]: { codec: dagPB, complete: false }
}

export const ethCodecs = {
  [dagEthHeader.name]: { codec: dagEthHeader, complete: true },
  [dagEthLog.name]: { codec: dagEthLog, complete: true },
  [dagEthLogTrie.name]: { codec: dagEthLogTrie, complete: true },
  [dagEthRct.name]: { codec: dagEthRct, complete: true },
  [dagEthRctTrie.name]: { codec: dagEthRctTrie, complete: true },
  [dagEthStateAccount.name]: { codec: dagEthStateAccount, complete: true },
  [dagEthStateTrie.name]: { codec: dagEthStateTrie, complete: true },
  [dagEthStorageTrie.name]: { codec: dagEthStorageTrie, complete: true },
  [dagEthTx.name]: { codec: dagEthTx, complete: true },
  [dagEthTxTrie.name]: { codec: dagEthTxTrie, complete: true },
  [dagEthUncles.name]: { codec: dagEthUncles, complete: true },
}
