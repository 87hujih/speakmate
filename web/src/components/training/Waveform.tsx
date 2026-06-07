interface WaveformProps {
  active?: boolean;
}

export function Waveform({ active = true }: WaveformProps) {
  return (
    <div className="flex h-12 flex-1 items-center justify-center gap-1">
      {Array.from({ length: 18 }).map((_, index) => {
        const height = index % 3 === 0 ? 34 : index % 2 === 0 ? 26 : 16;
        return (
          <i
            key={index}
            className="wave-bar w-1.5 rounded-full bg-gradient-to-b from-brand-blue to-brand-purple opacity-80"
            style={{ height, animationDelay: `${(index % 5) * 0.08}s` }}
            data-active={active}
          />
        );
      })}
    </div>
  );
}
