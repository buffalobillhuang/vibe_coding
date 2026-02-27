import React from 'react';
import { Museum } from '@/services/types';
import MuseumCard from './MuseumCard';

interface MuseumPanelProps {
  city: string;
  museums: Museum[];
  loading: boolean;
}

const MuseumPanel: React.FC<MuseumPanelProps> = ({ city, museums, loading }) => {
  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <p className="text-gray-500">正在加载博物馆数据...</p>
      </div>
    );
  }

  if (!museums || museums.length === 0) {
    return (
      <div className="flex items-center justify-center h-full p-8 text-center">
        <div>
          <h2 className="text-xl font-semibold text-gray-700 mb-2">未找到博物馆</h2>
          <p className="text-sm text-gray-500">我们还没有 {city} 的博物馆数据。</p>
        </div>
      </div>
    );
  }

  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold text-gray-900 mb-4">{city}</h1>
      <p className="text-sm text-gray-500 mb-6">共找到 {museums.length} 个博物馆</p>
      
      <div className="space-y-6">
        {museums.map((museum) => (
          <MuseumCard key={museum.id} museum={museum} />
        ))}
      </div>
    </div>
  );
};

export default MuseumPanel;
