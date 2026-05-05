import type { Metadata } from 'next';

type BuildMetadataInput = {
  path: `/${string}/`;
  title: string;
  description: string;
  image?: string;
};

const BASE_URL = 'https://moedah.com';
const DEFAULT_SITE_NAME = 'Moedah POS';
const DEFAULT_IMAGE = '/og-default.jpg';

export function buildMetadata(input: BuildMetadataInput): Metadata {
  const url = `${BASE_URL}${input.path}`;
  const imageUrl = `${BASE_URL}${input.image ?? DEFAULT_IMAGE}`;

  return {
    title: input.title,
    description: input.description,
    alternates: { canonical: url },
    openGraph: {
      type: 'website',
      url,
      title: input.title,
      description: input.description,
      siteName: DEFAULT_SITE_NAME,
      locale: 'id_ID',
      images: [{ url: imageUrl, width: 1200, height: 630, alt: input.title }],
    },
    twitter: {
      card: 'summary_large_image',
      title: input.title,
      description: input.description,
      images: [imageUrl],
    },
  };
}
