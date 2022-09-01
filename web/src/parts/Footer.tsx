import Container from '@components/Container';
import useStatus from '@hooks/status';
import LogoImage from '@images/logo.png';

export default function Footer() {
  const { data } = useStatus();
  return (
    <footer>
      <Container>
        <div className="flex flex-col items-center border-t border-slate-400/10 py-10 sm:flex-row-reverse sm:justify-between">
          <div>
            <div className="flex items-center font-bold text-lg">
              powered by
              <a
                href="https://github.com/samcm/checkpointz"
                className="flex items-center pl-1 hover:animate-pulse"
                aria-label="Checkpointz GitHub"
              >
                <span className="bg-clip-text font-extrabold text-lg text-transparent tracking-tighest bg-gradient-to-r from-rose-400 via-fuchsia-500 to-red-500">
                  Checkpoint
                </span>
                <img className="w-5 pl-1 pt-3" src={LogoImage} alt="checkpointz logo" />
              </a>
            </div>
            <div className="font-bold text-xs text-center sm:text-end text-gray-400">
              {data?.data?.version?.full}
            </div>
          </div>
          <div className="mt-6 lg:mt-0 flex items-center sm:mt-0">
            <a href="/" aria-label="Home" className="flex items-center">
              {data?.data?.brand_image_url && (
                <img src={data.data.brand_image_url} alt="logo" className="h-10 w-auto" />
              )}
              <span className="font-bold text-xl pl-2 text-gray-600">{data?.data?.brand_name}</span>
            </a>
          </div>
        </div>
      </Container>
    </footer>
  );
}
