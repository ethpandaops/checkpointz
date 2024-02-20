import { useState, useMemo } from 'react';

import {
  ExclamationTriangleIcon,
  ChevronDoubleRightIcon,
  InformationCircleIcon,
  CheckCircleIcon,
  XCircleIcon,
} from '@heroicons/react/24/outline';

import CopyToClipboard from '@components/CopyToClipboard';
import useStatus from '@hooks/status';

import GetStartedSelection, { ConsensusClient } from './GetStartedSelection';

export default function GetStarted() {
  const [client, setClient] = useState<ConsensusClient | undefined>(undefined);
  const { data } = useStatus({ refetchInterval: 60_000 });
  const publicURL = useMemo(() => {
    const defaultPublicURL = `${window.location.origin}${
      window.location.pathname === '/' ? '' : window.location.pathname
    }`;
    return `${data?.data?.public_url ?? defaultPublicURL}`;
  }, [data, client]);
  const lightModeWarning = useMemo(() => {
    if (data?.data?.operating_mode !== 'light') return;
    return (
      <div className="rounded-md bg-red-50 p-4 mt-5 shadow">
        <div className="flex">
          <div className="flex-shrink-0">
            <XCircleIcon className="h-5 w-5 text-red-500" aria-hidden="true" />
          </div>
          <div className="ml-3">
            <h3 className="text-sm font-semibold text-red-800">
              This Checkpointz instance is running in <span className="font-bold">light mode</span>
            </h3>
            <div className="mt-2 text-sm text-red-700">
              <p>
                A light mode Checkpointz instance can only be used for verifying! It can{' '}
                <span className="font-semibold">not</span> be used for beacon node checkpoint
                syncing. Please use a Checkpointz instance operating in full mode for checkpoint
                sync. This step is left visible for informational purposes.
              </p>
            </div>
          </div>
        </div>
      </div>
    );
  }, [data]);
  return (
    <>
      <GetStartedSelection onChange={setClient} />
      {client && (
        <div className="font-medium text-gray-700 sm:m-5">
          <div className="py-5">
            This guide covers the additional steps required to checkpoint sync a beacon node from
            another beacon node that you trust. This guide does not cover setting up an entire node
            from scratch.
          </div>
          <div className="rounded-md bg-yellow-50 p-4 shadow">
            <div className="flex">
              <div className="flex-shrink-0">
                <ExclamationTriangleIcon className="h-5 w-5 text-yellow-500" aria-hidden="true" />
              </div>
              <div className="ml-3">
                <h3 className="text-sm font-semibold text-yellow-800">Remember</h3>
                <div className="mt-2 text-sm text-yellow-700">
                  <p>
                    You should <span className="font-bold">always</span> verify that your beacon
                    node is synced correctly after doing a checkpoint sync. Refer to{' '}
                    <span className="font-bold">step 4</span> for more information.
                  </p>
                </div>
              </div>
            </div>
          </div>
          <div className="py-5">
            More reading on checkpoint sync:
            <ul className="mt-2 space-y-3 break-all">
              <li className="flex items-center sm:ml-5">
                <ChevronDoubleRightIcon className="hidden sm:inline h-4 w-4 text-fuchsia-500" />
                <a
                  href="https://www.symphonious.net/2022/05/21/checkpoint-sync-safety/"
                  className="underline"
                >
                  https://www.symphonious.net/2022/05/21/checkpoint-sync-safety/
                </a>
              </li>
              <li className="flex items-center sm:ml-5">
                <ChevronDoubleRightIcon className="hidden sm:inline h-4 w-4 text-fuchsia-500" />
                <a
                  href="https://ethereum.org/en/developers/docs/consensus-mechanisms/pos/weak-subjectivity/"
                  className="underline"
                >
                  https://ethereum.org/en/developers/docs/consensus-mechanisms/pos/weak-subjectivity/
                </a>
              </li>
              <li className="flex items-center sm:ml-5">
                <ChevronDoubleRightIcon className="hidden sm:inline h-4 w-4 text-fuchsia-500" />
                <a
                  href="https://blog.ethereum.org/2014/11/25/proof-stake-learned-love-weak-subjectivity/"
                  className="underline"
                >
                  https://blog.ethereum.org/2014/11/25/proof-stake-learned-love-weak-subjectivity/
                </a>
              </li>
              <li className="flex items-center sm:ml-5">
                <ChevronDoubleRightIcon className="hidden sm:inline h-4 w-4 text-fuchsia-500" />
                <a
                  href="https://notes.ethereum.org/@djrtwo/ws-sync-in-practice"
                  className="underline"
                >
                  https://notes.ethereum.org/@djrtwo/ws-sync-in-practice
                </a>
              </li>
            </ul>
          </div>
          <div className="relative mt-10">
            <div className="absolute inset-0 flex items-center" aria-hidden="true">
              <div className="w-full border-t border-gray-300" />
            </div>
            <div className="relative flex justify-start">
              <span className="bg-white pr-3 text-2xl font-bold text-gray-900">Step 1</span>
            </div>
          </div>
          {lightModeWarning}
          <div className="py-5">
            Note down the beacon endpoint you&apos;re planning to checkpoint sync from. This can be
            another beacon node you run, a beacon node that a friend runs, the endpoint of this
            Checkpointz instance, or any beacon node you trust.
          </div>
          <div className="pb-5">
            There is a maintained list of public hosted Checkpointz endpoints in this{' '}
            <a
              className="underline text-fuchsia-500 hover:text-fuchsia-600"
              href="https://github.com/eth-clients/checkpoint-sync-endpoints"
            >
              repository
            </a>
            .
          </div>
          <div className="rounded-md bg-blue-50 p-4 shadow">
            <div className="flex">
              <div className="flex-shrink-0">
                <InformationCircleIcon className="h-5 w-5 text-blue-400" aria-hidden="true" />
              </div>
              <div className="ml-3 flex-1 md:flex md:justify-between">
                <p className="text-sm text-blue-700">
                  The source beacon node must be for the same Ethereum network as your beacon node.
                </p>
              </div>
            </div>
          </div>
          {!lightModeWarning && (
            <div className="mt-5">
              <div className="rounded-md bg-green-50 p-4 shadow">
                <div className="flex">
                  <div className="flex-shrink-0">
                    <CheckCircleIcon className="h-5 w-5 text-green-400" aria-hidden="true" />
                  </div>
                  <div className="ml-3">
                    <p className="text-sm font-medium text-green-800">
                      The current Checkpointz instance endpoint will now be used for the rest of
                      this guide.
                    </p>
                  </div>
                </div>
              </div>
              <div className="flex rounded-md mt-5">
                <input
                  type="text"
                  value={publicURL}
                  name="endpoint"
                  id="endpoint"
                  disabled
                  className="p-2 w-full rounded-none rounded-l-lg border-gray-300 bg-gray-200 font-bold text-xl shadow-inner text-gray-800"
                />
                <span className="rounded-r-lg border min-w-max border-l-0 border-gray-300 bg-fuchsia-500 p-3 text-gray-100 text-xl">
                  <CopyToClipboard text={publicURL} inverted />
                </span>
              </div>
            </div>
          )}
          <div className="relative mt-10">
            <div className="absolute inset-0 flex items-center" aria-hidden="true">
              <div className="w-full border-t border-gray-300" />
            </div>
            <div className="relative flex justify-start">
              <span className="bg-white pr-3 text-2xl font-bold text-gray-900">Step 2</span>
            </div>
          </div>
          {lightModeWarning}
          <div className="mt-5">
            Add the checkpoint sync argument to your client.
            <div className="mt-5">{client.commandLine?.(publicURL)}</div>
          </div>
          <div className="relative mt-10">
            <div className="absolute inset-0 flex items-center" aria-hidden="true">
              <div className="w-full border-t border-gray-300" />
            </div>
            <div className="relative flex justify-start">
              <span className="bg-white pr-3 text-2xl font-bold text-gray-900">Step 3</span>
            </div>
          </div>
          {lightModeWarning}
          <div className="mt-5">
            Start your client. Once started, check your logs for details surrounding the checkpoint
            process.
            <div className="mt-5">{client.logCheck?.(publicURL)}</div>
          </div>
          <div className="relative mt-10">
            <div className="absolute inset-0 flex items-center" aria-hidden="true">
              <div className="w-full border-t border-gray-300" />
            </div>
            <div className="relative flex justify-start">
              <span className="bg-white pr-3 text-2xl font-bold text-gray-900">Step 4</span>
            </div>
          </div>
          <div className="mt-5">
            Validate that your node is on the expected chain. To do this we&apos;ll check the state
            root of the finalized checkpoint against another source.
          </div>
          <div className="mt-5">
            You will need to know the <span className="font-bold">IP</span> &{' '}
            <span className="font-bold">Port</span> of your beacon node.
            {client.defaultPort && (
              <span>
                The default port for {client.name} is{' '}
                <span className="font-mono bg-gray-100 p-1">{client.defaultPort}</span>
              </span>
            )}
          </div>
          <div className="mt-5 text-xl font-semibold">Obtaining slot and state root</div>
          <div className="mt-2 font-semibold">
            Option A
            <ol className="list-decimal font-normal">
              <li className="ml-10">
                Check your consensus client logs
                {client.name !== 'Not applicable' && client.logCheck?.(publicURL)}
              </li>
              <li className="ml-10">
                Find the <span className="font-mono bg-gray-100 p-1">slot</span> number.
              </li>
              <li className="ml-10">
                Find the <span className="font-mono bg-gray-100 p-1">state_root</span> value.
              </li>
            </ol>
          </div>
          <div className="mt-2 font-semibold">
            Option B
            <ol className="list-decimal font-normal">
              <li className="ml-10">
                Open{' '}
                <span className="font-mono bg-gray-100 p-1 break-all">
                  http://YOUR_NODE_IP:YOUR_NODE_PORT/eth/v1/beacon/headers/finalized
                </span>{' '}
                in your browser.
              </li>
              <li className="ml-10">
                Find the <span className="font-mono bg-gray-100 p-1">slot</span> number.
              </li>
              <li className="ml-10">
                Find the <span className="font-mono bg-gray-100 p-1">state_root</span> value.
              </li>
            </ol>
          </div>
          <div className="mt-2 font-semibold">
            Option C
            <ol className="list-decimal font-normal">
              <li className="ml-10">
                Install <span className="font-mono bg-gray-100 p-1">curl</span> and{' '}
                <span className="font-mono bg-gray-100 p-1">jq</span>.
              </li>
              <li className="ml-10">
                In a new terminal window run:
                <div className="bg-gray-100 rounded-lg grid">
                  <pre className="overflow-x-auto p-5">
                    curl -s http://YOUR_NODE_IP:YOUR_NODE_PORT/eth/v1/beacon/headers/finalized | jq
                    .&apos;data.header.message&apos;
                  </pre>
                </div>
              </li>
            </ol>
          </div>
          <div className="mt-5 text-xl font-semibold">Validate against a known trusted source</div>
          <div className="rounded-md bg-blue-50 p-4 shadow my-5">
            <div className="flex">
              <div className="flex-shrink-0">
                <InformationCircleIcon className="h-5 w-5 text-blue-400" aria-hidden="true" />
              </div>
              <div className="ml-3 flex-1 md:flex md:justify-between">
                <p className="text-sm text-blue-700">
                  The following references to{' '}
                  <span className="font-mono bg-blue-100 p-1">slot</span> and{' '}
                  <span className="font-mono bg-blue-100 p-1">state root</span> are the values from
                  the above{' '}
                  <span className="font-mono bg-blue-100 p-1">Obtaining slot and state root</span>{' '}
                  section.
                </p>
              </div>
            </div>
          </div>
          <div className="mt-2">
            You must verify the <span className="font-mono bg-gray-100 p-1">slot</span> and{' '}
            <span className="font-mono bg-gray-100 p-1">state root</span> against a known trusted
            source. This can be a friend, someone from the community that you know or any other
            source that you trust. There is a maintained list of public hosted checkpoint sync
            endpoints in this{' '}
            <a
              className="underline text-fuchsia-500 hover:text-fuchsia-600"
              href="https://github.com/eth-clients/checkpoint-sync-endpoints"
            >
              repository
            </a>
            , but it is recommended to use your own trusted source first.
          </div>
          <div className="mt-5">
            To verify your <span className="font-mono bg-gray-100 p-1">slot</span> and{' '}
            <span className="font-mono bg-gray-100 p-1">state root</span> against a Checkpointz
            instance;
            <ol className="list-decimal font-normal">
              <li className="ml-10">
                Open another Checkpointz instance website.{' '}
                <span className="font-semibold">
                  This must be different to this Checkpointz instance!
                </span>
              </li>
              <li className="ml-10">
                Check the historical finalized epoch boundaries table and search for the row that
                contains your <span className="font-mono bg-gray-100 p-1">slot</span> value.
              </li>
              <li className="ml-10">
                Make sure your <span className="font-mono bg-gray-100 p-1">state root</span>{' '}
                matches.
              </li>
            </ol>
          </div>
          <div className="mt-5">
            If it&apos;s a match, congratulations 🎉. If it&apos;s not a match you should start from
            scratch by wiping your beacon node and starting from the top.
          </div>
        </div>
      )}
    </>
  );
}
