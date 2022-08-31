import Footer from '@parts/Footer';
import Header from '@parts/Header';
import Hero from '@parts/Hero';
import PrimarySection from '@parts/PrimarySection';

export default function App() {
  return (
    <>
      <Header />
      <main>
        <Hero />
        <PrimarySection />
      </main>
      <Footer />
    </>
  );
}
