import { Fragment, useMemo, useState } from 'react';

import { Dialog, Transition } from '@headlessui/react';
import { InformationCircleIcon, XMarkIcon } from '@heroicons/react/24/outline';

import Container from '@components/Container';
import useStatus from '@hooks/status';
import LogoImage from '@images/logo.png';
import { getMajorityNetworkName } from '@utils';

export default function Header() {
  const { data, isLoading } = useStatus();
  const [open, setOpen] = useState(false);
  const majorityNetwork = useMemo(() => {
    return getMajorityNetworkName(Object.values(data?.data?.upstreams ?? {})) ?? 'unknown';
  }, [data]);
  const onClick = () => {
    setOpen(true);
  };
  return (
    <header className="relative z-50 pb-5 pt-2 sm:pb-7 lg:pt-7">
      <div className="hidden sm:block absolute top-0 left-1/2 transform -translate-x-1/2 bg-gradient-to-r from-rose-400 via-fuchsia-500 to-indigo-500 p-2 rounded-b-lg border-x-2 border-b-2 border-white/10 shadow-lg">
        <div className="flex items-center justify-center">
          <dt className="truncate text-sm tracking-tighter font-semibold text-gray-100">
            Network:
          </dt>
          <dd className="pl-1 underline capitalize text-sm font-bold tracking-tighter text-gray-100 hover:text-gray-200 hover:animate-pulse cursor-default">
            {majorityNetwork}
          </dd>
        </div>
        <div className="flex items-center justify-center">
          <dt className="truncate text-sm tracking-tighter font-semibold text-gray-100">
            Operation Mode:
          </dt>
          <dd className="pl-1 underline capitalize text-sm font-bold tracking-tighter text-gray-100 hover:text-gray-200 hover:animate-pulse cursor-default">
            {data?.data?.operating_mode ?? 'unknown'}
          </dd>
        </div>
      </div>
      <nav>
        <Container className="relative z-50 grid grid-cols-2 content-end py-2">
          <div className="relative z-10 flex items-center justify-self-start">
            <a href="/" aria-label="Home" className="flex items-center">
              {!data?.data?.brand_image_url && !data?.data?.brand_name && !isLoading ? (
                <>
                  <span className="bg-clip-text font-extrabold text-2xl text-transparent tracking-tighest bg-gradient-to-r from-rose-400 via-fuchsia-500 to-red-500">
                    Checkpoint
                  </span>
                  <img className="w-6 pl-1 pt-3" src={LogoImage} alt="checkpointz logo" />
                </>
              ) : (
                <>
                  {data?.data?.brand_image_url && (
                    <img src={data.data.brand_image_url} alt="logo" className="h-12 w-auto" />
                  )}
                  <span className="font-bold text-2xl pl-2 text-gray-600">
                    {data?.data?.brand_name}
                  </span>
                </>
              )}
            </a>
          </div>
          <div className="hidden sm:flex items-center justify-self-end">
            <span
              className="cursor-pointer text-gray-600 font-bold hover:text-gray-700 flex items-center"
              onClick={onClick}
            >
              <span className="pr-1">About</span>
              <InformationCircleIcon className="w-8 text-fuchsia-500 hover:text-fuchsia-600" />
            </span>
          </div>
          <div
            onClick={onClick}
            className="sm:hidden cursor-pointer absolute top-0 right-0 transform -translate-y-2 bg-gradient-to-r from-rose-400 via-fuchsia-500 to-indigo-500 p-2 rounded-bl-lg border-l-2 border-b-2 border-white/10 shadow-lg"
          >
            <div className="flex items-center justify-center">
              <dt className="truncate text-sm tracking-tighter font-semibold text-gray-100">
                Network:
              </dt>
              <dd className="pl-1 underline capitalize text-sm font-bold tracking-tighter text-gray-100 hover:text-gray-200 hover:animate-pulse cursor-default">
                {majorityNetwork}
              </dd>
            </div>
            <div className="flex items-center justify-center">
              <dt className="truncate text-sm tracking-tighter font-semibold text-gray-100">
                Operation Mode:
              </dt>
              <dd className="pl-1 underline capitalize text-sm font-bold tracking-tighter text-gray-100 hover:text-gray-200 hover:animate-pulse cursor-default">
                {data?.data?.operating_mode ?? 'unknown'}
              </dd>
            </div>
          </div>
        </Container>
      </nav>
      <Transition.Root show={open} as={Fragment}>
        <Dialog as="div" className="relative z-50" onClose={setOpen}>
          <Transition.Child
            as={Fragment}
            enter="ease-in-out duration-500"
            enterFrom="opacity-0"
            enterTo="opacity-100"
            leave="ease-in-out duration-500"
            leaveFrom="opacity-100"
            leaveTo="opacity-0"
          >
            <div className="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" />
          </Transition.Child>
          <div className="fixed inset-0" />

          <div className="fixed inset-0 overflow-hidden">
            <div className="absolute inset-0 overflow-hidden">
              <div className="pointer-events-none fixed inset-y-0 right-0 flex max-w-full pl-10 sm:pl-16">
                <Transition.Child
                  as={Fragment}
                  enter="transform transition ease-in-out duration-500 sm:duration-300"
                  enterFrom="translate-x-full"
                  enterTo="translate-x-0"
                  leave="transform transition ease-in-out duration-500 sm:duration-300"
                  leaveFrom="translate-x-0"
                  leaveTo="translate-x-full"
                >
                  <Dialog.Panel className="pointer-events-auto w-screen max-w-5xl">
                    <div className="flex h-full flex-col overflow-y-scroll bg-white py-6 shadow-xl">
                      <div className="px-4 sm:px-6">
                        <div className="flex items-start justify-between">
                          <Dialog.Title></Dialog.Title>
                          <div className="ml-3 flex h-7 items-center">
                            <button
                              type="button"
                              className="rounded-md bg-white text-gray-400 hover:text-gray-500"
                              onClick={() => setOpen(false)}
                            >
                              <span className="sr-only">Close panel</span>
                              <XMarkIcon className="h-6 w-6" aria-hidden="true" />
                            </button>
                          </div>
                        </div>
                      </div>
                      <div className="relative mt-6 flex-1">
                        <div className="relative pt-10 pb-20 sm:py-12">
                          <div className="absolute inset-x-0 -top-5 -bottom-14 overflow-hidden bg-indigo-50">
                            <div className="absolute inset-x-0 top-0 h-40 bg-gradient-to-b from-white" />
                            <div className="absolute inset-x-0 bottom-0 h-40 bg-gradient-to-t from-white" />
                          </div>
                          <Container className="relative">
                            <div className="mx-auto max-w-2xl lg:max-w-4xl lg:px-12">
                              <h1 className="font-display text-4xl font-bold tracking-tighter bg-clip-text text-transparent bg-gradient-to-r from-fuchsia-600 to-pink-600 sm:text-5xl lg:text-6xl">
                                An Ethereum beacon chain checkpoint sync provider
                              </h1>
                              <div className="mt-6 space-y-6 font-display text-xl sm:text-2xl tracking-tight text-gray-600">
                                <p>
                                  <span className="font-semibold">Checkpointz</span> exists to
                                  reduce the operational burden of running a checkpoint sync
                                  endpoint.
                                </p>
                                <p>
                                  If you&apos;d like to run your own instance of{' '}
                                  <span className="font-semibold">Checkpointz</span>, checkout out
                                  the{' '}
                                  <a
                                    className="underline text-fuchsia-500 hover:text-fuchsia-600"
                                    href="https://github.com/samcm/checkpointz"
                                  >
                                    Github repository
                                  </a>{' '}
                                  for instructions.
                                </p>
                              </div>
                            </div>
                          </Container>
                        </div>
                      </div>
                    </div>
                  </Dialog.Panel>
                </Transition.Child>
              </div>
            </div>
          </div>
        </Dialog>
      </Transition.Root>
    </header>
  );
}
