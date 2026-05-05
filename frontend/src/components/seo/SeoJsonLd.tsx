import { getSchemaByKey } from '@/lib/schema-pages';
import type { SeoPageKey } from '@/lib/seo-pages';

type SeoJsonLdProps = {
  pageKey: SeoPageKey;
};

export default function SeoJsonLd({ pageKey }: SeoJsonLdProps) {
  const schema = getSchemaByKey(pageKey);

  return (
    <script
      type="application/ld+json"
      dangerouslySetInnerHTML={{ __html: JSON.stringify(schema) }}
    />
  );
}
