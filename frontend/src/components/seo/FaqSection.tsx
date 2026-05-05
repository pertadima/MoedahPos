import type { SeoPageKey } from '@/lib/seo-pages';
import { faqByPage } from '@/lib/faq-pages';

type FaqSectionProps = {
  pageKey: SeoPageKey;
};

export default function FaqSection({ pageKey }: FaqSectionProps) {
  const items = faqByPage[pageKey];

  return (
    <section className="mt-10">
      <h2 className="text-2xl font-semibold text-gray-900">FAQ</h2>
      <div className="mt-4 space-y-4">
        {items.map(item => (
          <article key={item.q} className="rounded-lg border border-gray-200 p-4">
            <h3 className="font-semibold text-gray-900">{item.q}</h3>
            <p className="mt-1 text-gray-700">{item.a}</p>
          </article>
        ))}
      </div>
    </section>
  );
}
