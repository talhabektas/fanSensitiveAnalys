// Takım logoları ve bilgileri
import gsLogo from '../assets/images/gs.png';
import fbLogo from '../assets/images/fb.jpg';
import bjkLogo from '../assets/images/jk.jpg';
import tsLogo from '../assets/images/ts.jpg';


export const teamLogos = {
  'galatasaray': gsLogo,
  'fenerbahce': fbLogo,
  'besiktas': bjkLogo,
  'trabzonspor': tsLogo,
};

export const teamNames = {
  'galatasaray': 'Galatasaray',
  'fenerbahce': 'Fenerbahçe',
  'besiktas': 'Beşiktaş',
  'trabzonspor': 'Trabzonspor',
};

export const teamColors = {
  'galatasaray': {
    primary: '#FFD700',
    secondary: '#CC1010',
    background: 'linear-gradient(135deg, #FFD700 0%, #FFA500 100%)'
  },
  'fenerbahce': {
    primary: '#FFFF00',
    secondary: '#000080',
    background: 'linear-gradient(135deg, #FFFF00 0%, #FFD700 100%)'
  },
  'besiktas': {
    primary: '#000000',
    secondary: '#FFFFFF',
    background: 'linear-gradient(135deg, #000000 0%, #333333 100%)'
  },
  'trabzonspor': {
    primary: '#800020',
    secondary: '#87CEEB',
    background: 'linear-gradient(135deg, #800020 0%, #A0522D 100%)'
  },
};

// Takım isminden logo al
export const getTeamLogo = (teamName) => {
  if (!teamName) return null;
  
  // Türkçe karakterleri çevir ve temizle
  const normalizedName = teamName
    .toLowerCase()
    .replace(/ğ/g, 'g')
    .replace(/ç/g, 'c')
    .replace(/ş/g, 's')
    .replace(/ı/g, 'i')
    .replace(/ö/g, 'o')
    .replace(/ü/g, 'u')
    .replace(/[^a-z]/g, '');
  
  if (normalizedName.includes('galatasaray') || normalizedName.includes('gs')) {
    return teamLogos.galatasaray;
  }
  if (normalizedName.includes('fenerbahce') || normalizedName.includes('fb')) {
    return teamLogos.fenerbahce;
  }
  if (normalizedName.includes('besiktas') || normalizedName.includes('bjk') || normalizedName.includes('jk')) {
    return teamLogos.besiktas;
  }
  if (normalizedName.includes('trabzonspor') || normalizedName.includes('ts')) {
    return teamLogos.trabzonspor;
  }
  return null;
};

// Takım isminden renk al
export const getTeamColors = (teamName) => {
  if (!teamName) return { primary: '#6B7280', secondary: '#9CA3AF', background: 'linear-gradient(135deg, #6B7280 0%, #9CA3AF 100%)' };
  
  // Türkçe karakterleri çevir ve temizle
  const normalizedName = teamName
    .toLowerCase()
    .replace(/ğ/g, 'g')
    .replace(/ç/g, 'c')
    .replace(/ş/g, 's')
    .replace(/ı/g, 'i')
    .replace(/ö/g, 'o')
    .replace(/ü/g, 'u')
    .replace(/[^a-z]/g, '');
  
  if (normalizedName.includes('galatasaray') || normalizedName.includes('gs')) {
    return teamColors.galatasaray;
  }
  if (normalizedName.includes('fenerbahce') || normalizedName.includes('fb')) {
    return teamColors.fenerbahce;
  }
  if (normalizedName.includes('besiktas') || normalizedName.includes('bjk') || normalizedName.includes('jk')) {
    return teamColors.besiktas;
  }
  if (normalizedName.includes('trabzonspor') || normalizedName.includes('ts')) {
    return teamColors.trabzonspor;
  }
  
  return { primary: '#6B7280', secondary: '#9CA3AF', background: 'linear-gradient(135deg, #6B7280 0%, #9CA3AF 100%)' };
};

// TeamLogo komponenti
export const TeamLogo = ({ teamName, size = 32, className = '' }) => {
  const logo = getTeamLogo(teamName);
  const colors = getTeamColors(teamName);
  
  if (!logo) {
    return (
      <div 
        className={`inline-flex items-center justify-center rounded-full ${className}`}
        style={{ 
          width: size, 
          height: size, 
          background: colors.background,
          minWidth: size 
        }}
      >
        <span 
          className="text-white font-bold text-xs"
          style={{ fontSize: Math.max(8, size / 4) }}
        >
          {teamName ? teamName.substring(0, 2).toUpperCase() : '?'}
        </span>
      </div>
    );
  }
  
  return (
    <img 
      src={logo} 
      alt={teamName}
      className={`rounded-full object-cover ${className}`}
      style={{ width: size, height: size, minWidth: size }}
    />
  );
};