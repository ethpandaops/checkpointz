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
INFO Loaded checkpoint block and state       state_root: 0x854ca984298e6a0d9fc098b4e37f1b28727e545a8e4d3106188fda3587d14cdb, block_root: 0x91e4f5129bc54f7284f1690e00803db360aeb63d34610d6995e1145bd01c0d92, slot: 529024, service: beacon
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
info: Download completed
info: Initializing beacon state from anchor state slot=529024, epoch=16532, stateRoot=0x854ca984298e6a0d9fc098b4e37f1b28727e545a8e4d3106188fda3587d14cdb`}
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
Downloading checkpoint state               restUrl=${publicURL} checkpointSlot=529024
Writing checkpoint state                   stateRoot=91e4f512
Writing checkpoint block                   blockRoot=13b4cc5f blck="(slot: 529024, proposer_index: 1466, parent_root: \\"0b019674\\", state_root: \\"854ca984\\")"`}
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
level=info msg="requesting ${publicURL}/eth/v2/debug/beacon/states/finalized"
level=info msg="requesting ${publicURL}/eth/v2/beacon/blocks/0x91e4f5129bc54f7284f1690e00803db360aeb63d34610d6995e1145bd01c0d92"
level=info msg="BeaconState slot=529024, Block slot=529024"
level=info msg="BeaconState htr=0x854ca984298e6a0d9fc098b4e37f1b28727e545a8e4d3106188fda3587d14cdbd, Block state_root=0x854ca984298e6a0d9fc098b4e37f1b28727e545a8e4d3106188fda3587d14cdb"`}
        </pre>
      </div>
    ),
  },
  {
    name: 'Teku',
    image: TekuImage,
    imageClassName: 'bg-[#ff6e42] p-1',
    description: 'Teku is a Java-based Ethereum 2.0 client developed by ConsenSys.',
    defaultPort: 5051,
    commandLine: (publicURL: string) => (
      <div className="bg-gray-100 rounded-lg grid">
        <pre className="overflow-x-auto p-5">--checkpoint-sync-url={publicURL}</pre>
      </div>
    ),
    logCheck: (publicURL: string) => (
      <div className="bg-gray-100 rounded-lg grid">
        <pre className="overflow-x-auto p-5">
          {`INFO  - Loading initial state from ${publicURL}
INFO  - Loaded initial state at epoch 16532 (state root = 0x854ca984298e6a0d9fc098b4e37f1b28727e545a8e4d3106188fda3587d14cdb, block root = 0x91e4f5129bc54f7284f1690e00803db360aeb63d34610d6995e1145bd01c0d92, block slot = 529024).  Please ensure that the supplied initial state corresponds to the latest finalized block as of the start of epoch 16532 (slot 529024)."`}
        </pre>
      </div>
    ),
  },
];

export default function GetStartedSelection(props: {
  onChange?: (client?: ConsensusClient) => void;
}) {
  const [selected, setSelected] = useState<ConsensusClient | undefined>(undefined);

  const onSelect = (client?: ConsensusClient) => {
    setSelected(client);
    props.onChange?.(client);
  };

  if (selected)
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
      <div className="flex flex-col gap-3">
        {clients.map((client) => (
          <div
            key={client.name}
            onClick={() => onSelect(client)}
            className="relative block cursor-pointer rounded-lg border bg-white px-6 py-4 shadow-sm focus:outline-none sm:flex sm:justify-between border-gray-300 hover:border-fuchsia-500 hover:ring-2 hover:ring-fuchsia-500"
          >
            <span className="flex items-center">
              <span className="flex flex-col text-sm">
                <span className="font-medium text-gray-900 flex items-center">
                  {client.image && (
                    <img
                      src={client.image}
                      alt={client.name}
                      className={clsx(client.imageClassName, 'w-6 border rounded shadow')}
                    />
                  )}
                  <span className="text-lg font-bold pl-1">{client.name}</span>
                </span>
                <span className="text-gray-500">{client.description}</span>
              </span>
            </span>
            <span
              className="pointer-events-none absolute -inset-px rounded-lg border hover:border-2'"
              aria-hidden="true"
            />
          </div>
        ))}
      </div>
    </>
  );
}
