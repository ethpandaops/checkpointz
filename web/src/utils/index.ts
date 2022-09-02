import { APIUpstream } from '@types';

export function truncateHash(hash?: string): string {
  if (!hash) return '';
  return hash.substring(0, 8) + '...' + hash.substring(hash.length - 6, hash.length);
}

// https://github.com/ChainSafe/web3.js/blob/1.x/packages/web3-utils/src/index.js#L166
export function hexToAscii(hex?: string) {
  if (!hex) return '';
  let str = '';
  let i = 2;
  const l = hex.length;
  for (; i < l; i += 2) {
    str += String.fromCharCode(parseInt(hex.slice(i, i + 2), 16));
  }
  return str;
}

export function stringToHexColour(str: string): string {
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    hash = str.charCodeAt(i) + ((hash << 5) - hash);
    hash = hash & hash;
  }
  let color = '#';
  for (let i = 0; i < 3; i++) {
    const value = (hash >> (i * 8)) & 255;
    color += value.toString(16).substring(-2);
  }
  // make sure the hex color is always 6 characters
  if (color.length < 7) {
    color += Array.from({ length: 7 - color.length }, () => '0').join('');
  } else if (color.length > 7) {
    color = color.substring(0, 7);
  }
  return color;
}

export function getMajorityNetworkName(upstreams: APIUpstream[]): string | undefined {
  const networkMap = upstreams.reduce<Record<string, number>>((acc, upstream) => {
    if (!upstream.network_name) return acc;
    if (!acc[upstream.network_name]) acc[upstream.network_name] = 0;
    acc[upstream.network_name] += 1;
    return acc;
  }, {});
  return Object.entries(networkMap).sort((a, b) => b[1] - a[1])[0]?.[0];
}
