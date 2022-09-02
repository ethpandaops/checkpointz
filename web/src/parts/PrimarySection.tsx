import { useEffect, useState } from 'react';

import { Tab } from '@headlessui/react';
import clsx from 'clsx';

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
      className="relative overflow-hidden shadow-inner bg-gradient-to-r from-rose-400 via-fuchsia-500 to-indigo-500"
    >
      <Tab.Group as="div" className="gap-y-6 py-10" vertical={tabOrientation === 'vertical'}>
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
            <Tab.Panels className="block sm:flex justify-center">
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
