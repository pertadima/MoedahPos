import type { SeoPageKey } from '@/lib/seo-pages';
import { faqByPage } from '@/lib/faq-pages';

type FaqSectionProps = {
  pageKey: SeoPageKey;
};

export default function FaqSection({ pageKey }: FaqSectionProps) {
  const items = faqByPage[pageKey];

  return (
    <section className="mt-10">
      <h2 className="text-2xl font-semibold text-[#0b3f6f]">FAQ</h2>
      <div className="mt-4 space-y-4">
        {items.map(item => (
          <article
            key={item.q}
            className="rounded-xl border border-[#0884F6]/15 bg-white p-4 shadow-sm shadow-[#0884F6]/5"
          >
            <h3 className="font-semibold text-[#0f3f6a]">{item.q}</h3>
            <p className="mt-1 text-[#325b7f]">{item.a}</p>
          </article>
        ))}
      </div>
    </section>
  );
}
