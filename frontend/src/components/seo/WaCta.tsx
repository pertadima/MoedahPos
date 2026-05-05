type WaCtaProps = {
  title: string;
  description: string;
  buttonText: string;
  waNumber?: string;
  waMessage?: string;
};

const DEFAULT_WA_NUMBER = '6200000000000';

export default function WaCta({
  title,
  description,
  buttonText,
  waNumber = DEFAULT_WA_NUMBER,
  waMessage = 'Halo Moedah POS, saya ingin demo.',
}: WaCtaProps) {
  const resolvedNumber = process.env.NEXT_PUBLIC_WA_NUMBER ?? waNumber;
  const href = `https://wa.me/${resolvedNumber}?text=${encodeURIComponent(waMessage)}`;

  return (
    <section className="mt-10 rounded-xl border border-emerald-200 bg-emerald-50 p-6">
      <h2 className="text-2xl font-semibold text-emerald-900">{title}</h2>
      <p className="mt-2 text-emerald-800">{description}</p>
      <a
        href={href}
        target="_blank"
        rel="noopener noreferrer"
        className="mt-4 inline-flex rounded-lg bg-emerald-600 px-4 py-2 font-semibold text-white hover:bg-emerald-700"
      >
        {buttonText}
      </a>
    </section>
  );
}
