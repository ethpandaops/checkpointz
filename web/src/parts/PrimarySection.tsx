import { useEffect, useState } from 'react';

import { Tab } from '@headlessui/react';
import clsx from 'clsx';

import useStatus from '@hooks/status';
import Checkpoints from '@parts/checkpoints/Checkpoints';
import Upstream from '@parts/upstream/Upstream';

const Sections = [
  {
    title: 'Checkpoints',
    children: <Checkpoints />,
  },
  {
    title: 'Upstream',
    children: <Upstream />,
  },
];

export default function Section() {
  const { data } = useStatus();
  const [tabOrientation, setTabOrientation] = useState('horizontal');

  useEffect(() => {
    const lgMediaQuery = window.matchMedia('(min-width: 1024px)');

    function onMediaQueryChange({ matches }: { matches: boolean }) {
      setTabOrientation(matches ? 'vertical' : 'horizontal');
    }

    onMediaQueryChange(lgMediaQuery);
    lgMediaQuery.addEventListener('change', onMediaQueryChange);

    return () => {
      lgMediaQuery.removeEventListener('change', onMediaQueryChange);
    };
  }, []);

  return (
    <section
      id="sections"
      className="relative overflow-hidden shadow-inner bg-scooter-200 bg-gradient-to-r from-rose-400 via-fuchsia-500 to-indigo-500"
    >
      {data?.data?.operating_mode === 'light' && (
        <div className="absolute w-full">
          <div className="overflow-hidden h-screen">
            <div
              className="bg-gradient-to-r from-pink-500 via-red-500 to-yellow-500 origin-top float-right mt-3 mr-3 w-32 text-center"
              style={{ transform: 'translateX(50%) rotate(45deg)' }}
            >
              <div
                className="text-sm text-gray-100 font-bold"
                title="Checkpointz instance is running in light operation mode"
              >
                Light
              </div>
            </div>
          </div>
        </div>
      )}
      <Tab.Group
        as="div"
        className="grid grid-cols-1 items-center gap-y-6 py-10"
        vertical={tabOrientation === 'vertical'}
      >
        {({ selectedIndex }) => (
          <>
            <div className="-mx-4 px-1 flex overflow-x-auto sm:overflow-visible pb-0">
              <Tab.List className="relative z-10 flex gap-x-4 whitespace-nowrap mx-auto px-0">
                {Sections.map((section, featureIndex) => (
                  <div
                    key={section.title}
                    className={clsx(
                      'group relative rounded-full py-1 px-4',
                      selectedIndex === featureIndex ? 'bg-white' : 'hover:bg-white/10',
                    )}
                  >
                    <h3>
                      <Tab
                        className={clsx(
                          'font-display text-base sm:text-lg focus:outline-none font-semibold',
                          selectedIndex === featureIndex
                            ? 'text-fuchsia-500'
                            : 'text-fuchsia-100 hover:text-white',
                        )}
                      >
                        <span className="absolute inset-0 rounded-full" />
                        {section.title}
                      </Tab>
                    </h3>
                  </div>
                ))}
              </Tab.List>
            </div>
            <Tab.Panels className="lg:col-span-7">
              {Sections.map((section) => (
                <Tab.Panel key={section.title} unmount={false}>
                  <div className="relative">
                    <div className="" />
                    {section.children}
                  </div>
                </Tab.Panel>
              ))}
            </Tab.Panels>
          </>
        )}
      </Tab.Group>
    </section>
  );
}
