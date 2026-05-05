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
    <section className="mt-10 rounded-2xl border border-[#0884F6]/20 bg-gradient-to-br from-[#0884F6]/10 via-white to-[#FFA724]/10 p-6">
      <h2 className="text-2xl font-semibold text-[#0b3f6f]">{title}</h2>
      <p className="mt-2 text-[#1f4f7a]">{description}</p>
      <a
        href={href}
        target="_blank"
        rel="noopener noreferrer"
        className="mt-4 inline-flex rounded-xl bg-[#0884F6] px-4 py-2 font-semibold text-white shadow-md shadow-[#0884F6]/20 hover:bg-[#0776dc]"
      >
        {buttonText}
      </a>
    </section>
  );
}
