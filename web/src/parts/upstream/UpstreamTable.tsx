import { useState, useMemo } from 'react';

import { ExclamationTriangleIcon } from '@heroicons/react/24/outline';
import clsx from 'clsx';

import CopyToClipboard from '@components/CopyToClipboard';
import Tooltip from '@components/Tooltip';
import { APIUpstream } from '@types';
import { truncateHash, stringToHexColour, getMajorityNetworkName } from '@utils';

export default function UpstreamTable(props: { upstreams: APIUpstream[] }) {
  const majorityNetwork = useMemo(() => {
    return getMajorityNetworkName(props.upstreams) ?? 'unknown';
  }, [props.upstreams]);
  const [search, setSearch] = useState('');
  const filteredUpstreams = useMemo(() => {
    if (!search) return props.upstreams;
    return props.upstreams.filter((upstream) => {
      return (
        upstream.name.toLowerCase().includes(search.toLowerCase()) ||
        (upstream.healthy ? 'healthy' : 'unhealthy').includes(search.toLowerCase()) ||
        upstream.finality?.finalized?.root.toLowerCase().includes(search.toLowerCase()) ||
        upstream.finality?.finalized?.epoch.toLowerCase().includes(search.toLowerCase()) ||
        upstream.finality?.current_justified?.root.toLowerCase().includes(search.toLowerCase()) ||
        upstream.finality?.current_justified?.epoch.toLowerCase().includes(search.toLowerCase())
      );
    });
  }, [props.upstreams, search]);

  const onSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => setSearch(e.target.value);
  return (
    <div className="px-4 sm:px-6 lg:px-8">
      <div className="mt-4 sm:flex sm:justify-end">
        <div className="hidden sm:flex flex-auto mr-4 p-2 rounded-md">
          <p className="mt-2 text-lg text-gray-100 drop-shadow-lg font-semibold self-end">
            Upstream beacon nodes of this Checkpointz server.
          </p>
        </div>
        <div className="bg-white/20 p-2 rounded-md">
          <label
            htmlFor="email"
            className="block text-lg drop-shadow-lg font-semibold text-gray-100"
          >
            Search
          </label>
          <div className="mt-1">
            <input
              type="email"
              name="email"
              id="email"
              className="shadow-lg block w-full sm:text-lg border-gray-300 rounded-md p-1"
              onChange={onSearchChange}
            />
          </div>
        </div>
      </div>
      <div className="mt-4 flex flex-col">
        <div className="-my-2 -mx-4 overflow-x-auto sm:-mx-6 lg:-mx-8">
          <div className="inline-block min-w-full py-2 align-middle md:px-6 lg:px-8">
            <div className="overflow-hidden shadow ring-1 ring-black ring-opacity-5 md:rounded-lg">
              <table className="min-w-full divide-y divide-gray-300">
                <thead className="bg-white/20">
                  <tr>
                    <th
                      scope="col"
                      className="whitespace-nowrap drop-shadow-lg pl-2 sm:px-2 py-3.5 text-sm sm:text-base text-left font-bold text-gray-100 sm:pl-6"
                    >
                      Name
                    </th>
                    <th
                      scope="col"
                      className="whitespace-nowrap drop-shadow-lg sm:px-2 py-3.5 text-left text-sm sm:text-base font-bold text-gray-100"
                    >
                      Status
                    </th>
                    <th
                      scope="col"
                      className="drop-shadow-lg whitespace-nowrap sm:px-2 py-3.5 text-left text-sm sm:text-base font-bold text-gray-100"
                    >
                      Network
                    </th>
                    <th
                      scope="col"
                      className="whitespace-nowrap drop-shadow-lg sm:px-2 py-3.5 text-left text-sm sm:text-base font-bold text-gray-100"
                    >
                      Finalized Epoch
                    </th>
                    <th
                      scope="col"
                      className="hidden sm:table-cell drop-shadow-lg whitespace-nowrap sm:px-2 py-3.5 text-left text-sm sm:text-base font-bold text-gray-100"
                    >
                      Finalized Block Root
                    </th>
                    <th
                      scope="col"
                      className="hidden lg:table-cell drop-shadow-lg whitespace-nowrap sm:px-2 py-3.5 text-left text-sm sm:text-base font-bold text-gray-100"
                    >
                      Justified Epoch
                    </th>
                    <th
                      scope="col"
                      className="hidden lg:table-cell drop-shadow-lg whitespace-nowrap sm:px-2 py-3.5 text-left text-sm sm:text-base font-bold text-gray-100"
                    >
                      Justified Block Root
                    </th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-200 bg-white/10">
                  {filteredUpstreams.map((upstream) => (
                    <tr key={upstream.name}>
                      <td className="whitespace-nowrap drop-shadow-lg pl-2 sm:px-2 py-2 font-semibold text-sm sm:text-base text-gray-100 sm:pl-6">
                        {upstream.name}
                      </td>
                      <td className="whitespace-nowrap drop-shadow-lg py-2 sm:px-2 text-sm sm:text-base font-semibold text-gray-100">
                        <span
                          className={clsx(
                            upstream.healthy
                              ? 'text-green-800 bg-green-100'
                              : 'text-red-800 bg-red-100',
                            'flex-shrink-0 inline-block px-2 py-0.5 text-xs font-semibold rounded-full',
                          )}
                        >
                          {upstream.healthy ? 'Healthy' : 'Unhealthy'}
                        </span>
                      </td>
                      <td className="drop-shadow-lg capitalize whitespace-nowrap sm:px-2 py-2 text-sm sm:text-base font-semibold text-gray-100">
                        {upstream.network_name ?? 'unknown'}
                        {upstream.network_name !== majorityNetwork && (
                          <ExclamationTriangleIcon className="h-4 w-4 text-yellow-400 inline ml-1" />
                        )}
                      </td>
                      <td className="whitespace-nowrap drop-shadow-lg sm:px-2 py-2 text-sm sm:text-base font-semibold text-gray-100">
                        {upstream.finality?.finalized?.epoch ?? ''}
                      </td>
                      <td className="hidden sm:table-cell whitespace-nowrap sm:px-2 py-2 text-sm sm:text-base font-semibold text-gray-100 font-mono">
                        {upstream.finality?.finalized?.root ? (
                          <span className="flex items-center">
                            <span
                              className="w-4 h-4 inline-block rounded mr-1 mt-1 border border-fuchsia-300"
                              style={{
                                backgroundColor: stringToHexColour(
                                  upstream.finality.finalized.root,
                                ),
                              }}
                            ></span>
                            <span className="font-mono cursor-pointer flex">
                              <Tooltip content={upstream.finality.finalized.root}>
                                <span className="relative top-1 group transition duration-300">
                                  {truncateHash(upstream.finality.finalized.root)}
                                  <span className="relative -top-0.5 block max-w-0 group-hover:max-w-full transition-all duration-500 h-0.5 bg-gray-100"></span>
                                </span>
                              </Tooltip>
                              <CopyToClipboard text={upstream.finality.finalized.root} inverted />
                            </span>
                          </span>
                        ) : (
                          ''
                        )}
                      </td>
                      <td className="hidden lg:table-cell drop-shadow-lg whitespace-nowrap sm:px-2 py-2 text-sm sm:text-base font-semibold text-gray-100">
                        {upstream.finality?.current_justified?.epoch ?? ''}
                      </td>
                      <td className="hidden lg:table-cell drop-shadow-lg whitespace-nowrap sm:px-2 py-2 text-sm sm:text-base font-semibold text-gray-100 font-mono">
                        {upstream.finality?.current_justified?.root ? (
                          <span className="flex items-center">
                            <span
                              className="w-4 h-4 inline-block rounded mr-1 mt-1 border border-fuchsia-300"
                              style={{
                                backgroundColor: stringToHexColour(
                                  upstream.finality.current_justified.root,
                                ),
                              }}
                            ></span>
                            <span className="font-mono cursor-pointer flex">
                              <Tooltip content={upstream.finality.current_justified.root}>
                                <span className="relative top-1 group transition duration-300">
                                  {truncateHash(upstream.finality.current_justified.root)}
                                  <span className="relative -top-0.5 block max-w-0 group-hover:max-w-full transition-all duration-500 h-0.5 bg-gray-100"></span>
                                </span>
                              </Tooltip>
                              <CopyToClipboard
                                text={upstream.finality.current_justified.root}
                                inverted
                              />
                            </span>
                          </span>
                        ) : (
                          ''
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
