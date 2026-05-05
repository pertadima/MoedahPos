import type { MetadataRoute } from 'next';

export default function sitemap(): MetadataRoute.Sitemap {
  const baseUrl = 'https://moedah.com';
  const now = new Date();

  return [
    {
      url: `${baseUrl}/aplikasi-pos-indonesia/`,
      lastModified: now,
      changeFrequency: 'weekly',
      priority: 0.9,
    },
    {
      url: `${baseUrl}/aplikasi-pos/`,
      lastModified: now,
      changeFrequency: 'weekly',
      priority: 0.9,
    },
    {
      url: `${baseUrl}/aplikasi-kasir-digital/`,
      lastModified: now,
      changeFrequency: 'weekly',
      priority: 0.8,
    },
    {
      url: `${baseUrl}/aplikasi-pos-restoran/`,
      lastModified: now,
      changeFrequency: 'weekly',
      priority: 0.8,
    },
    {
      url: `${baseUrl}/aplikasi-pos-umkm/`,
      lastModified: now,
      changeFrequency: 'weekly',
      priority: 0.8,
    },
    {
      url: `${baseUrl}/moka-pos-alternatif/`,
      lastModified: now,
      changeFrequency: 'weekly',
      priority: 0.7,
    },
    {
      url: `${baseUrl}/harga-aplikasi-pos/`,
      lastModified: now,
      changeFrequency: 'weekly',
      priority: 0.8,
    },
  ];
}
