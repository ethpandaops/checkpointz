import { useState } from 'react';

import Tippy, { TippyProps } from '@tippyjs/react/headless';
import clsx from 'clsx';

export default function Tooltip(props: TippyProps & { children?: JSX.Element }) {
  const [visible, setVisible] = useState(false);
  const { children, ...tippyProps } = props;
  return (
    <Tippy
      interactive={true}
      render={(attrs, content) => (
        <div
          className={clsx(
            'bg-gray-800 text-gray-100 px-3 py-1 rounded-lg shadow-lg',
            visible ? '' : 'animate-fade hidden',
          )}
          {...attrs}
        >
          <span className="font-mono font-bold text-sm sm:text-base">
            {content}
            {props.content}
          </span>
        </div>
      )}
      appendTo={document.body}
      trigger="click"
      delay={500}
      interactiveBorder={20}
      placement="top"
      onShow={() => setVisible(true)}
      onHide={() => setVisible(false)}
      {...tippyProps}
    >
      {children}
    </Tippy>
  );
}
