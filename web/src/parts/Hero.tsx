import { InformationCircleIcon } from '@heroicons/react/20/solid';

import CircleBackground from '@components/CircleBackground';
import CopyToClipboard from '@components/CopyToClipboard';
import Tooltip from '@components/Tooltip';
import useStatus from '@hooks/status';
import FlagImage from '@images/flag.png';
import GetStartedSlideout from '@parts/get-started/GetStartedSlideout';
import { truncateHash } from '@utils';

export default function Status() {
  const { data } = useStatus({ refetchInterval: 60_000 });
  return (
    <div className="relative pb-10">
      <div className="z-10 relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="max-w-6xl mx-auto">
          <h1 className="font-display text-2xl font-bold text-center tracking-tighter bg-clip-text text-transparent bg-gradient-to-r from-rose-400 via-fuchsia-500 to-indigo-500 sm:text-4xl lg:text-5xl">
            An Ethereum beacon chain checkpoint sync provider
          </h1>
        </div>
        <div className="max-w-4xl mx-auto pt-4 sm:pt-10">
          <dl className="rounded-lg  sm:grid sm:grid-cols-2">
            <div className="flex flex-col p-2 sm:p-6 text-center ">
              <dd className="order-2 mt-2 text-lg leading-6 items-baseline font-medium text-gray-500 flex self-center">
                <span className="text-base font-semibold text-gray-600 whitespace-nowrap">
                  Epoch:
                </span>
                {data?.data.finality?.finalized?.epoch ? (
                  <span className="pl-1 text-base text-gray-500">
                    {data?.data.finality.finalized.epoch}
                  </span>
                ) : (
                  <div className="animate-pulse bg-slate-200 rounded grow w-50 text-transparent">
                    0000000
                  </div>
                )}
              </dd>
              <dd className="order-2 mt-2 text-lg leading-6 items-baseline font-medium text-gray-500 flex self-center">
                <span className="text-base font-semibold text-gray-600 whitespace-nowrap">
                  Block Root:
                </span>
                {data?.data.finality?.finalized?.root ? (
                  <span className="pl-1 text-base text-gray-500 flex font-mono items-start cursor-pointer">
                    <Tooltip content={data.data.finality.finalized.root}>
                      <span className="relative group transition duration-300">
                        {truncateHash(data.data.finality.finalized.root)}
                        <span className="relative -top-0.5 block max-w-0 group-hover:max-w-full transition-all duration-500 h-0.5 bg-fuchsia-400"></span>
                      </span>
                    </Tooltip>
                    <CopyToClipboard text={data.data.finality.finalized.root} />
                  </span>
                ) : (
                  <div className="animate-pulse bg-slate-200 rounded grow w-50 text-transparent">
                    0x000000...000000
                  </div>
                )}
              </dd>
              <dd className="order-1 text-xl tracking-tight font-bold text-fuchsia-500">
                Latest Finalized
                <Tooltip content="The current finalized checkpoint being served by this Checkpointz instance">
                  <InformationCircleIcon className="w-5 inline pl-0.5 align-text-top cursor-pointer text-fuchsia-500 hover:text-fuchsia-600" />
                </Tooltip>
              </dd>
            </div>
            <div className="flex flex-col p-2 sm:p-6 text-center">
              <dd className="order-2 mt-2 text-lg leading-6 items-baseline font-medium text-gray-500 flex self-center">
                <span className="text-base font-semibold text-gray-600 whitespace-nowrap">
                  Epoch:
                </span>
                {data?.data.finality?.current_justified?.epoch ? (
                  <span className="pl-1 text-base text-gray-500">
                    {data?.data.finality.current_justified.epoch}
                  </span>
                ) : (
                  <div className="animate-pulse bg-slate-200 rounded grow w-50 text-transparent">
                    0000000
                  </div>
                )}
              </dd>
              <dd className="order-2 mt-2 text-lg leading-6 items-baseline font-medium text-gray-500 flex self-center">
                <span className="text-base font-semibold text-gray-600 whitespace-nowrap">
                  Block Root:
                </span>
                {data?.data.finality?.current_justified?.root ? (
                  <span className="pl-1 text-base text-gray-500 flex font-mono items-start cursor-pointer">
                    <Tooltip content={data.data.finality.current_justified.root}>
                      <span className="relative group transition duration-300">
                        {truncateHash(data.data.finality.current_justified.root)}
                        <span className="relative -top-0.5 block max-w-0 group-hover:max-w-full transition-all duration-500 h-0.5 bg-fuchsia-400"></span>
                      </span>
                    </Tooltip>
                    <CopyToClipboard text={data.data.finality.current_justified.root} />
                  </span>
                ) : (
                  <div className="animate-pulse bg-slate-200 rounded grow w-50 text-transparent">
                    0x000000...000000
                  </div>
                )}
              </dd>
              <dd className="order-1 text-xl tracking-tight font-bold text-fuchsia-500">
                Latest Justified
                <Tooltip content="The latest justified checkpoint known by this Checkpointz instance">
                  <InformationCircleIcon className="w-5 inline pl-0.5 align-text-top cursor-pointer text-fuchsia-500 hover:text-fuchsia-600" />
                </Tooltip>
              </dd>
            </div>
          </dl>
        </div>
      </div>
      <div>
        <div className="flex justify-center mt-4">
          <GetStartedSlideout />
        </div>
      </div>
      <div className="absolute top-1/2 top-40 sm:top-40 left-20 -translate-y-1/2 left-1/2 -translate-x-1/2 opacity-30 lg:opacity-50">
        <CircleBackground color="#d946ef" width={200} height={200} className="animate-spin-slow" />
      </div>
      <div className="absolute top-1/2 top-40 sm:top-40 left-20 -translate-y-1/2 left-1/2 -translate-x-1/2 opacity-30 lg:opacity-50">
        <CircleBackground
          color="#d946ef"
          width={250}
          height={250}
          className="animate-spin-slower"
        />
      </div>
      <div className="z-10 absolute top-1/2 top-40 sm:top-40 left-20 -translate-y-1/2 left-1/2 -translate-x-1/2 opacity-5 lg:opacity-20">
        <img src={FlagImage} alt="flag" width={75} height={75} />
      </div>
      <div className="-z-50 relative">
        <div className="absolute inset-x-0 -top-48 -bottom-12 overflow-hidden bg-fuchsia-50">
          <div className="absolute inset-x-0 top-0 h-40 bg-gradient-to-b from-white" />
          <div className="absolute inset-x-0 bottom-0 h-40 bg-gradient-to-t from-white" />
        </div>
      </div>
    </div>
  );
}
