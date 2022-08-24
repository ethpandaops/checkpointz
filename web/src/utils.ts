
export function truncateHash(hash?: string): string {
  if (!hash) return '';
  return hash.substring(0, 8) + '...' + hash.substring(hash.length-6, hash.length);
};

// https://github.com/ChainSafe/web3.js/blob/1.x/packages/web3-utils/src/index.js#L166
export function hexToAscii(hex?: string) {
  if (!hex) return '';
  let str = '';
  var i = 2, l = hex.length;
  for (; i < l; i += 2) {
    str += String.fromCharCode(parseInt(hex.slice(i, i + 2), 16));
  }
  return str;
};
