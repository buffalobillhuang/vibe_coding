import axios from 'axios';
import { MuseumResponse, MuseumList } from './types';

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8000/api';

export const fetchMuseums = async (city: string): Promise<MuseumList> => {
  try {
    const response = await axios.get<MuseumResponse>(`${API_URL}/museums/${encodeURIComponent(city)}`);
    return response.data.museums;
  } catch (error) {
    if (axios.isAxiosError(error)) {
      console.error(`Failed to load museums for ${city}:`, error.message);
      throw new Error(`Failed to load museum data: ${error.message}`);
    }
    throw new Error('An unexpected error occurred');
  }
};

export const API_ENDPOINTS = {
  museums: (city: string) => `${API_URL}/museums/${encodeURIComponent(city)}`,
};
