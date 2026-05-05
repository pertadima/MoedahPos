import { seoPages, type SeoPageKey } from '@/lib/seo-pages';
import { faqByPage } from '@/lib/faq-pages';

const BASE_URL = 'https://moedah.com';

type SchemaGraph = {
  '@context': 'https://schema.org';
  '@graph': Array<Record<string, unknown>>;
};

const orgNode = {
  '@type': 'Organization',
  '@id': `${BASE_URL}/#organization`,
  name: 'Moedah POS',
  url: BASE_URL,
  logo: `${BASE_URL}/logo.png`,
};

function breadcrumbNode(name: string, path: string) {
  return {
    '@type': 'BreadcrumbList',
    itemListElement: [
      { '@type': 'ListItem', position: 1, name: 'Beranda', item: `${BASE_URL}/` },
      { '@type': 'ListItem', position: 2, name, item: `${BASE_URL}${path}` },
    ],
  };
}

function softwareNode(name: string, description: string, path: string) {
  return {
    '@type': 'SoftwareApplication',
    name,
    applicationCategory: 'BusinessApplication',
    operatingSystem: 'Web',
    inLanguage: 'id-ID',
    description,
    url: `${BASE_URL}${path}`,
    provider: { '@id': `${BASE_URL}/#organization` },
  };
}

function webPageNode(name: string, description: string, path: string) {
  return {
    '@type': 'WebPage',
    name,
    inLanguage: 'id-ID',
    description,
    url: `${BASE_URL}${path}`,
    isPartOf: { '@id': `${BASE_URL}/#website` },
  };
}

function faqNode(items: Array<{ q: string; a: string }>) {
  return {
    '@type': 'FAQPage',
    mainEntity: items.map(item => ({
      '@type': 'Question',
      name: item.q,
      acceptedAnswer: { '@type': 'Answer', text: item.a },
    })),
  };
}

const softwareKeys: SeoPageKey[] = [
  'aplikasi-pos',
  'aplikasi-kasir-digital',
  'aplikasi-pos-restoran',
  'aplikasi-pos-umkm',
];

export function getSchemaByKey(key: SeoPageKey): SchemaGraph {
  const page = seoPages[key];
  const isSoftware = softwareKeys.includes(key);

  return {
    '@context': 'https://schema.org',
    '@graph': [
      orgNode,
      isSoftware
        ? softwareNode(page.title, page.description, page.path)
        : webPageNode(page.title, page.description, page.path),
      breadcrumbNode(page.title.replace(' | Moedah POS', ''), page.path),
      faqNode(faqByPage[key]),
    ],
  };
}
