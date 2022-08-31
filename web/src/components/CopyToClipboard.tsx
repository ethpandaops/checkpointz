import { ClipboardDocumentCheckIcon } from '@heroicons/react/24/outline';
import { useState } from 'react';
import clsx from 'clsx';

export default function CopyToClipboard(props: { text: string; inverted?: boolean; className?: string; }) {
  const [copied, setCopied] = useState(false);
  const onClick = () => {
    navigator.clipboard.writeText(props.text);
    setCopied(true);
  };
  return (<button
    className={clsx(
      'align-bottom group',
      copied
        ? 'animate-fade'
        : '',
      props.className,
    )}
  >
    <ClipboardDocumentCheckIcon
      onClick={onClick}
      className={clsx(
        'h-6 w-6 transition hover:rotate-[-4deg]',
        props.inverted
          ? 'stroke-gray-50 hover:stroke-gray-100'
          : 'stroke-fuchsia-500 hover:stroke-fuchsia-600'
      )}
    />
  </button>);
}
