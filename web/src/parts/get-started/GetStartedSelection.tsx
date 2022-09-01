import { useState } from 'react';

import { RadioGroup } from '@headlessui/react';
import { ExclamationTriangleIcon } from '@heroicons/react/24/outline';
import clsx from 'clsx';

import LighthouseImage from '@images/lighthouse.svg';
import LodestarImage from '@images/lodestar.png';
import NimbusImage from '@images/nimbus.png';
import PrysmImage from '@images/prysm.png';
import TekuImage from '@images/teku.svg';

export type ConsensusClient = {
  name: string;
  image?: string;
  imageClassName?: string;
  description: string;
  endpointPathSuffix?: string;
  commandLine?: (publicURL: string) => JSX.Element;
  logCheck?: (publicURL: string) => JSX.Element;
  defaultPort?: number;
};

const clients: ConsensusClient[] = [
  {
    name: 'Not applicable',
    description: "Choose this option if you don't want client specific details.",
    commandLine: () => (
      <div className="rounded-md bg-yellow-50 p-4 shadow">
        <div className="flex">
          <div className="flex-shrink-0">
            <ExclamationTriangleIcon className="h-5 w-5 text-yellow-500" aria-hidden="true" />
          </div>
          <div className="ml-3">
            <h3 className="text-sm font-semibold text-yellow-800">No Consensus client set</h3>
            <div className="mt-2 text-sm text-yellow-700">
              <p>
                Refer to your client&apos;s documentation for more information on how to checkpoint
                sync.
              </p>
            </div>
          </div>
        </div>
      </div>
    ),
    logCheck: () => (
      <div className="rounded-md bg-yellow-50 p-4 shadow">
        <div className="flex">
          <div className="flex-shrink-0">
            <ExclamationTriangleIcon className="h-5 w-5 text-yellow-500" aria-hidden="true" />
          </div>
          <div className="ml-3">
            <h3 className="text-sm font-semibold text-yellow-800">No Consensus client set</h3>
            <div className="mt-2 text-sm text-yellow-700">
              <p>
                Refer to your client&apos;s documentation for more information on how to checkpoint
                sync.
              </p>
            </div>
          </div>
        </div>
      </div>
    ),
  },
  {
    name: 'Lighthouse',
    image: LighthouseImage,
    imageClassName: 'bg-[#6563ff] p-1',
    description:
      'Lighthouse is an open-source Ethereum consensus client, written in Rust and maintained by Sigma Prime.',
    defaultPort: 5052,
    commandLine: (publicURL: string) => (
      <div className="bg-gray-100 rounded-lg grid">
        <pre className="overflow-x-auto p-5">--checkpoint-sync-url={publicURL}</pre>
      </div>
    ),
    logCheck: (publicURL: string) => (
      <div className="bg-gray-100 rounded-lg grid">
        <pre className="overflow-x-auto p-5">
          {`INFO Starting checkpoint sync                remote_url: ${publicURL}, service: beacon
INFO Loaded checkpoint block and state       state_root: 0xb6e8c25393411f252775f82b4907298572003ac37acf9422dd2500b5c368a08d, block_root: 0xefe19ea0d99bf45d50d4302f6bbc3feb2c1ec46f8d6e112594ec86b9581596ae, slot: 3480832, service: beacon
`}
        </pre>
      </div>
    ),
  },
  {
    name: 'Lodestar',
    image: LodestarImage,
    description:
      'Lodestar is a TypeScript implementation of the Ethereum Consensus specification developed by ChainSafe Systems.',
    defaultPort: 9596,
    commandLine: (publicURL: string) => (
      <div className="bg-gray-100 rounded-lg grid">
        <pre className="overflow-x-auto p-5">--checkpointSyncUrl={publicURL}</pre>
      </div>
    ),
    logCheck: (publicURL: string) => (
      <div className="bg-gray-100 rounded-lg grid">
        <pre className="overflow-x-auto p-5">
          {`info: Fetching weak subjectivity state weakSubjectivityServerUrl=${publicURL}
info: Download completed`}
        </pre>
      </div>
    ),
  },
  {
    name: 'Nimbus',
    image: NimbusImage,
    imageClassName: 'bg-[#f39200] p-1',
    description:
      'Nimbus is an extremely efficient Ethereum consensus layer client implementation developed by Status Research & Development.',
    defaultPort: 5052,
    commandLine: () => (
      <div>
        Requires the <span className="font-mono bg-gray-100 p-1">trustedNodeSync</span> command to
        be run before the beacon is launched. Read more here:{' '}
        <a className="underline" href="https://nimbus.guide/trusted-node-sync.html">
          https://nimbus.guide/trusted-node-sync.html
        </a>
      </div>
    ),
    logCheck: (publicURL: string) => (
      <div className="bg-gray-100 rounded-lg grid">
        <pre className="overflow-x-auto p-5">
          {`Starting trusted node sync                 databaseDir=/data/consensus/db restUrl=${publicURL} blockId=finalized backfill=false reindex=false
Downloading checkpoint block               restUrl=${publicURL} blockId=finalized
Downloading checkpoint state               restUrl=${publicURL} checkpointSlot=370528
Writing checkpoint state
Writing checkpoint block
Database initialized, historical blocks will be backfilled when starting the node`}
        </pre>
      </div>
    ),
  },
  {
    name: 'Prysm',
    image: PrysmImage,
    imageClassName: 'bg-[#22292f]',
    description:
      'Prysm is an Ethereum proof-of-stake client written in Go developed by Prysmatic Labs.',
    defaultPort: 3500,
    commandLine: (publicURL: string) => (
      <div className="bg-gray-100 rounded-lg grid">
        <pre className="overflow-x-auto p-5">
          {`--checkpoint-sync-url=${publicURL}
--genesis-beacon-api-url=${publicURL}`}
        </pre>
      </div>
    ),
    logCheck: (publicURL: string) => (
      <div className="bg-gray-100 rounded-lg grid">
        <pre className="overflow-x-auto p-5">
          {`level=info msg="requesting ${publicURL}/eth/v2/debug/beacon/states/genesis"
level=info msg="requesting ${publicURL}/eth/v2/debug/beacon/states/finalized"`}
        </pre>
      </div>
    ),
  },
  {
    name: 'Teku',
    image: TekuImage,
    imageClassName: 'bg-[#ff6e42] p-1',
    description: 'Teku is a Java-based Ethereum 2.0 client developed by ConsenSys.',
    endpointPathSuffix: '/eth/v2/debug/beacon/states/finalized',
    defaultPort: 5051,
    commandLine: (publicURL: string) => (
      <div className="bg-gray-100 rounded-lg grid">
        <pre className="overflow-x-auto p-5">--initial-state={publicURL}</pre>
      </div>
    ),
    logCheck: (publicURL: string) => (
      <div className="bg-gray-100 rounded-lg grid">
        <pre className="overflow-x-auto p-5">
          {`INFO  - Loading initial state from ${publicURL}
INFO  - Loaded initial state at epoch 11348 (state root = 0x08dab651bd667b166a0c99b7a21ee455f4f9fadfce0e37dbcee490f5ec841477, block root = 0xa5bd8b3eaadd81f892f120219f3bcee6565a37d045bf0cee4c4023a51def430c, block slot = 363136).  Please ensure that the supplied initial state corresponds to the latest finalized block as of the start of epoch 11348 (slot 363136)."`}
        </pre>
      </div>
    ),
  },
];

export default function GetStartedSelection(props: {
  onChange?: (client?: ConsensusClient) => void;
}) {
  const [selected, setSelected] = useState<ConsensusClient>(clients[0]);
  const [isSelected, setIsSelected] = useState<boolean>(false);

  const onSelect = (client?: ConsensusClient) => {
    props.onChange?.(client);
    setIsSelected(Boolean(client));
  };

  if (isSelected)
    return (
      <div className="flex items-center">
        <span className="pr-1 font-bold text-gray-500">Consensus client: </span>
        <span className="inline-flex items-center rounded-full bg-fuchsia-100 py-1 pl-2.5 pr-1 text-sm font-medium text-fuchsia-700">
          {selected.image && (
            <img
              src={selected.image}
              alt={selected.name}
              className={clsx(selected.imageClassName, 'w-6 border rounded shadow')}
            />
          )}
          <span className="text-lg font-bold pl-1">{selected.name}</span>
          <button
            type="button"
            className="ml-0.5 inline-flex h-4 w-4 flex-shrink-0 items-center justify-center rounded-full text-fuchsia-400 hover:bg-fuchsia-200 hover:text-fuchsia-500 focus:bg-fuchsia-500 focus:text-white focus:outline-none"
            onClick={() => onSelect(undefined)}
          >
            <span className="sr-only">Remove</span>
            <svg className="h-2 w-2" stroke="currentColor" fill="none" viewBox="0 0 8 8">
              <path strokeLinecap="round" strokeWidth="1.5" d="M1 1l6 6m0-6L1 7" />
            </svg>
          </button>
        </span>
      </div>
    );

  return (
    <>
      <h1 className="text-2xl font-bold m-4 text-gray-800">
        Which Ethereum consensus client are you using?
      </h1>
      <RadioGroup value={selected} onChange={setSelected}>
        <RadioGroup.Label className="sr-only">Consensus client</RadioGroup.Label>
        <div className="space-y-4">
          {clients.map((client) => (
            <RadioGroup.Option
              key={client.name}
              value={client}
              className={({ checked, active }) =>
                clsx(
                  checked ? 'border-transparent' : 'border-gray-300',
                  active ? 'border-fuchsia-500 ring-2 ring-fuchsia-500' : '',
                  'relative block cursor-pointer rounded-lg border bg-white px-6 py-4 shadow-sm focus:outline-none sm:flex sm:justify-between',
                )
              }
            >
              {({ active, checked }) => (
                <>
                  <span className="flex items-center">
                    <span className="flex flex-col text-sm">
                      <RadioGroup.Label
                        as="span"
                        className="font-medium text-gray-900 flex items-center"
                      >
                        {client.image && (
                          <img
                            src={client.image}
                            alt={client.name}
                            className={clsx(client.imageClassName, 'w-6 border rounded shadow')}
                          />
                        )}
                        <span className="text-lg font-bold pl-1">{client.name}</span>
                      </RadioGroup.Label>
                      <RadioGroup.Description as="span" className="text-gray-500">
                        {client.description}
                      </RadioGroup.Description>
                    </span>
                  </span>
                  <span
                    className={clsx(
                      active ? 'border' : 'border-2',
                      checked ? 'border-fuchsia-500' : 'border-transparent',
                      'pointer-events-none absolute -inset-px rounded-lg',
                    )}
                    aria-hidden="true"
                  />
                </>
              )}
            </RadioGroup.Option>
          ))}
        </div>
      </RadioGroup>
      <div className="flex justify-end mt-5">
        <button
          type="button"
          className="inline-flex items-center rounded-md border border-transparent bg-fuchsia-500 px-10 py-3 text-base font-medium text-white shadow-sm hover:bg-fuchsia-600 focus:outline-none focus:ring-2 focus:ring-fuchsia-500 focus:ring-offset-2"
          onClick={() => onSelect(selected)}
        >
          <span className="font-bold text-lg">Continue</span>
        </button>
      </div>
    </>
  );
}
