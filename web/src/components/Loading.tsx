import clsx from 'clsx';

import LogoImage from '@images/logo.png';

export default function Loading({ className, ...props }: { className?: string }) {
  return (
    <img
      className={clsx('w-20 animate-pulse opacity-80', className)}
      alt="Loading..."
      {...props}
      src={LogoImage}
    />
  );
}
