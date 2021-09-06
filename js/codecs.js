import * as dagCBOR from '@ipld/dag-cbor'
import * as dagJSON from '@ipld/dag-json'
import * as dagPB from '@ipld/dag-pb'
import * as dagJOSE from 'dag-jose'

export const codecs = {
  [dagCBOR.name]: { codec: dagCBOR, complete: true },
  [dagJSON.name]: { codec: dagJSON, complete: true },
  [dagPB.name]: { codec: dagPB, complete: false },
  [dagJOSE.name]: { codec: dagJOSE, complete: false }
}
