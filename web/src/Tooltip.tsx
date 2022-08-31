import { CopyToClipboard } from 'react-copy-to-clipboard';

export default function Tooltip(props: { text: string; tooltipText?: string; clipboardText?: string }) {
  return (
    <CopyToClipboard text={props.clipboardText ?? ''} data-tip={props.tooltipText}>
      <button className="group relative inline-block text-checkpointz underline hover:text-red-500 duration-300">
        {props.text}
      </button>
    </CopyToClipboard>
  );
}
