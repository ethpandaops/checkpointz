import clsx from 'clsx';

export default function Container({ className, children, ...props }: { className?: string; children: JSX.Element | JSX.Element[]; }) {
  return (
    <div
      className={clsx('mx-auto max-w-7xl px-4 sm:px-6 lg:px-8', className)}
      {...props}
    >{children}</div>
  )
}
