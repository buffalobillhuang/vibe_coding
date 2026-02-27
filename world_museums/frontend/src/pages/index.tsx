import React, { useState } from 'react';
import type { NextPage } from 'next';
import Head from 'next/head';
import { MuseumList } from '@/services/types';
import { fetchMuseums } from '@/services/api';
import ChinaMap from '@/components/ChinaMap';
import MuseumPanel from '@/components/MuseumPanel';

const Home: NextPage = () => {
  const [selectedCity, setSelectedCity] = useState<string | null>(null);
  const [museums, setMuseums] = useState<MuseumList>([]);
  const [loading, setLoading] = useState(false);

  console.log('Home component rendering');

  const handleCitySelect = async (city: string) => {
    console.log('handleCitySelect called with:', city);
    setSelectedCity(city);
    setLoading(true);
    try {
      const museumData = await fetchMuseums(city);
      setMuseums(museumData);
    } catch (error) {
      console.error('Error loading museums:', error);
      setMuseums([]);
    } finally {
      setLoading(false);
    }
  };

  return (
    <>
      <Head>
        <title>World Museums - China Map</title>
        <meta name="description" content="Explore museums across China on an interactive map" />
        <link rel="icon" href="/favicon.ico" />
      </Head>

      <div className="flex flex-col h-screen w-full bg-gray-100 overflow-hidden">
        {/* Left Panel: China Map - 80% width */}
        <div className="w-[80%] h-full relative shadow-lg">
          <ChinaMap onCitySelect={handleCitySelect} />
        </div>

        {/* Right Panel: Museum Details - 20% width */}
        <div className="w-[20%] h-full overflow-y-auto bg-white border-l shadow-xl">
          {!selectedCity ? (
            <EmptyState />
          ) : (
            <MuseumPanel city={selectedCity} museums={museums} loading={loading} />
          )}
        </div>
      </div>
    </>
  );
};

function EmptyState() {
  return (
    <div className="flex items-center justify-center h-full p-8 text-center">
      <div>
        <h2 className="text-xl font-semibold text-gray-700 mb-2">选择城市</h2>
        <p className="text-sm text-gray-500">点击地图上的城市查看博物馆信息</p>
      </div>
    </div>
  );
}

export default Home;
